// Copyright 2018 CoreOS Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package torcx

import (
	"bufio"
	"bytes"
	"context"
	_ "crypto/sha512" // used by go-digest
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/euank/gotmpl"
	"github.com/northbright/ctx/ctxcopy"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/clearsign"
)

var (
	errNilRemote          = errors.New("nil Remote")
	errNilRemotesCache    = errors.New("nil RemotesCache")
	errEmptyUsrMountpoint = errors.New("empty USR mountpoint")
	errEmptyTemplateURL   = errors.New("empty remote URL template")
	errEmptyLocation      = errors.New("empty location")
)

// evaluateURL evaluates the URL template for a remote
// and performs variables substitution sourcing values from
// `/etc/os-release`.
func (r *Remote) evaluateURL(usrMountpoint string) (*url.URL, error) {
	if r == nil {
		return nil, errNilRemote
	}
	if usrMountpoint == "" {
		return nil, errEmptyUsrMountpoint
	}
	if r.TemplateURL == "" {
		return nil, errEmptyTemplateURL
	}

	if !needSubstitution(r.TemplateURL) {
		return url.Parse(r.TemplateURL)
	}

	osReleasePath := VendorOsReleasePath(usrMountpoint)
	fp, err := os.Open(osReleasePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %s", osReleasePath)
	}
	defer fp.Close()
	osMeta, err := parseOsRelease(fp)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", osReleasePath)
	}
	osMeta["COREOS_USR"] = usrMountpoint

	templateVars := map[string]string{
		"COREOS_BOARD": osMeta["COREOS_BOARD"],
		"COREOS_USR":   osMeta["COREOS_USR"],
		"ID":           osMeta["ID"],
		"VERSION_ID":   osMeta["VERSION_ID"],
	}
	urlRaw, err := gotmpl.TemplateString(r.TemplateURL, gotmpl.MapLookup(templateVars))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to evaluate template %s", r.TemplateURL)
	}

	return url.Parse(urlRaw)
}

// contentsURL returns the full evaluated URL to the remote contents manifest.
func (r *Remote) contentsURL(usrMountpoint string) (*url.URL, error) {
	manifestName, err := url.Parse("torcx_remote_contents.json.asc")
	if err != nil {
		return nil, err
	}
	baseURL, err := r.evaluateURL(usrMountpoint)
	if err != nil {
		return nil, err
	}
	fullURL := baseURL.ResolveReference(manifestName)
	return fullURL, nil
}

// loadKeyrings loads all keyrings referenced by a remote manifest.
// `baseDir` is used as the path prefix to find the keyrings by filename.
func (r *Remote) loadKeyrings(baseDir string) ([]openpgp.KeyRing, error) {
	if baseDir == "" {
		return nil, errors.New("empty base directory")
	}
	if r == nil {
		return nil, errors.New("nil Remote")
	}
	if r.TemplateURL == "" {
		return nil, errors.New("empty remote URL template")
	}

	keyrings := []openpgp.KeyRing{}
	for _, k := range r.ArmoredKeys {
		path := filepath.Join(baseDir, k)
		fp, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer fp.Close()
		el, err := openpgp.ReadArmoredKeyRing(fp)
		if err != nil {
			return nil, err
		}
		keyrings = append(keyrings, el)
	}

	return keyrings, nil
}

// needSubstitution checks whether a URL template contains any
// variables that need to be evaluated.
func needSubstitution(template string) bool {
	emptyMap := gotmpl.MapLookup(map[string]string{})
	_, err := gotmpl.TemplateString(template, emptyMap)
	needSub := err != nil
	return needSub
}

// parseOsRelease is the parser for os-release.
func parseOsRelease(rd io.Reader) (map[string]string, error) {
	meta := map[string]string{}

	sc := bufio.NewScanner(rd)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == "" {
			continue
		}
		value := strings.Trim(parts[1], `"`)
		if value == "" {
			continue
		}
		meta[parts[0]] = value
	}
	if sc.Err() != nil {
		return meta, sc.Err()
	}
	return meta, nil
}

// RemotesCache holds a temporary cache for images/references in the store
type RemotesCache struct {
	Configs       map[string]Remote
	Contents      map[string]RemoteContents
	Paths         map[string]string
	UsrMountpoint string
}

// NewRemotesCache constructs a new RemotesCache
func NewRemotesCache(ctx context.Context, usrMountpoint string, baseDirs []string, remotesFilter []string) (*RemotesCache, error) {
	rc := RemotesCache{
		Configs:       map[string]Remote{},
		Contents:      map[string]RemoteContents{},
		Paths:         map[string]string{},
		UsrMountpoint: usrMountpoint,
	}

	// Process all remote base directories and cache all remotes found.
	for _, dir := range baseDirs {
		glob := filepath.Join(dir, "*", "remote.json")
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}
		quotedDir := regexp.QuoteMeta(dir)
		re, err := regexp.Compile(fmt.Sprintf(`^%s/(.*)/remote\.json$`, quotedDir))
		if err != nil {
			return nil, err
		}
		for _, remote := range matches {
			groups := re.FindStringSubmatch(remote)
			if len(groups) != 2 {
				return nil, errors.Errorf("non-unique matches: %s", groups)
			}
			if groups[1] == "" {
				continue
			}
			name := groups[1]
			rc.Paths[name] = remote
		}
	}

	// Only keep relevant remotes for this cache.
	filtered := map[string]string{}
	for _, name := range remotesFilter {
		if path, ok := rc.Paths[name]; ok && path != "" {
			filtered[name] = path
		}
	}
	if len(remotesFilter) > 0 {
		rc.Paths = filtered
	}

	// Download and verify remote manifests.
	for name, path := range rc.Paths {
		fp, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer fp.Close()
		bufrd := bufio.NewReader(fp)
		var jm RemoteManifestV0JSON
		if err := json.NewDecoder(bufrd).Decode(&jm); err != nil {
			return nil, errors.Wrapf(err, "failed to decode %s", name)
		}
		if jm.Kind != RemoteManifestV0K {
			return nil, errors.Errorf("invalid manifest kind: %s", jm.Kind)
		}
		remote := RemoteFromJSONV0(jm.Value)
		rc.Configs[name] = remote
		url, err := remote.contentsURL(rc.UsrMountpoint)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to evaluate URL for %s", name)
		}
		keyrings, err := remote.loadKeyrings(filepath.Dir(path))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load keyrings for %s", name)
		}
		var manifest string
		switch url.Scheme {
		case "https":
			tries := 0
			for {
				tries++
				tmpManifest, err := fetchManifest(ctx, url.String())
				ctxErr := ctx.Err()
				if err == nil && ctxErr == nil {
					manifest = tmpManifest
					break
				}
				if ctxErr != nil {
					return nil, ctxErr
				}
				logrus.WithFields(logrus.Fields{
					"attempt": tries,
					"name":    name,
					"error":   err,
				}).Error("failed to fetch contents manifest")
				time.Sleep(8 * time.Second)
			}
		case "file":
			path := strings.TrimPrefix(url.String(), "file://")
			b, err := ioutil.ReadFile(filepath.Clean(path))
			if err != nil {
				return nil, errors.Wrapf(err, "failed to fetch contents manifest for %s", name)
			}
			manifest = string(b)
		default:
			return nil, errors.Errorf("unsupported scheme %s", url.Scheme)
		}

		unwrapped, err := verifyManifest(name, manifest, keyrings)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to verify contents manifest for %s", name)
		}
		contents, err := decodeContents(unwrapped)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode contents for %s", name)
		}
		rc.Contents[name] = *contents

		logrus.WithFields(logrus.Fields{
			"name": name,
			"path": path,
			"url":  url,
		}).Debug("remote verified")
	}

	// Length sanity check.
	if len(rc.Paths) != len(rc.Configs) || len(rc.Paths) != len(rc.Contents) {
		return nil, errors.Errorf("length mismatch, %d vs %d vs %d", len(rc.Paths), len(rc.Configs), len(rc.Contents))
	}

	return &rc, nil
}

// CheckAvailable checks if a given Image is available in the configured remote.
// On success, it returns the full evaluated base URL for the remote and
// the relative image location.
func (rc *RemotesCache) CheckAvailable(im Image) (*url.URL, *url.URL, string, error) {
	if im.Remote == "" {
		return nil, nil, "", nil
	}
	if rc == nil {
		return nil, nil, "", errors.New("nil RemotesCache")
	}

	contents, ok := rc.Contents[im.Remote]
	if !ok {
		return nil, nil, "", errors.Errorf("manifest for remote %s not found: %s", im.Remote, rc)
	}
	config, ok := rc.Configs[im.Remote]
	if !ok {
		return nil, nil, "", errors.Errorf("manifest for remote %s not found: %s", im.Remote, rc)
	}
	baseURL, err := config.evaluateURL(rc.UsrMountpoint)
	if err != nil {
		return nil, nil, "", errors.Wrapf(err, "failed to evaluate URL for %s", im.Remote)
	}
	location, hash, err := contents.CheckAvailable(im)
	if err != nil {
		return nil, nil, "", errors.Wrapf(err, "inspecting remote %s", im.Remote)
	}
	if location == nil {
		return nil, nil, "", nil
	}

	return baseURL, location, hash, nil
}

func verifyManifest(manifestName string, manifest string, keyrings []openpgp.KeyRing) (string, error) {
	if manifest == "" {
		return "", errors.New("empty manifest")
	}
	if len(keyrings) <= 0 {
		logrus.WithFields(logrus.Fields{
			"name": manifestName,
		}).Warn("no keys to verify manifest")
	}

	signedBlock, trailer := clearsign.Decode([]byte(manifest))
	if signedBlock == nil {
		if len(keyrings) == 0 {
			logrus.WithFields(logrus.Fields{
				"name": manifestName,
			}).Warn("unsigned manifest and no keys to verify it")
			return manifest, nil
		}
		return "", errors.New("no signed manifest detected")
	}
	if len(trailer) != 0 {
		return "", errors.New("trailing data after signed manifest")
	}
	if signedBlock.ArmoredSignature == nil {
		return "", errors.New("no clearsign data to verify")
	}
	if len(signedBlock.Plaintext) <= 0 {
		return "", errors.New("no plaintext to verify")
	}

	for _, kr := range keyrings {
		if _, err := openpgp.CheckDetachedSignature(kr, bytes.NewReader(signedBlock.Bytes), signedBlock.ArmoredSignature.Body); err == nil {
			return string(signedBlock.Plaintext), nil
		}
	}

	return "", errors.New("unable to verify contents manifest")
}

func fetchManifest(ctx context.Context, urlRaw string) (string, error) {
	var manifest bytes.Buffer
	req, err := http.NewRequest("GET", urlRaw, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	buf := make([]byte, 32*1024)
	if err := ctxcopy.Copy(ctx, &manifest, resp.Body, buf); err != nil {
		return "", err
	}
	return manifest.String(), nil
}

func decodeContents(manifest string) (*RemoteContents, error) {
	rd := strings.NewReader(manifest)
	bufrd := bufio.NewReader(rd)
	var container kindValueJSON
	if err := json.NewDecoder(bufrd).Decode(&container); err != nil {
		return nil, errors.Wrap(err, "failed to decode manifest")
	}

	switch container.Kind {
	case RemoteContentsV1K:
		manifest := RemoteContentsV1JSON{
			Kind: container.Kind,
		}
		if err := json.Unmarshal(container.Value, &manifest.Value); err != nil {
			return nil, err
		}
		value := RemoteContentsFromJSONV1(manifest.Value)
		return &value, nil
	}

	return nil, errors.Errorf("invalid manifest kind: %s", container.Kind)
}

// CheckAvailable checks if a given Image is available in the configured remote.
// On success, it returns its location (anchored at `base_url`).
func (rcs *RemoteContents) CheckAvailable(im Image) (*url.URL, string, error) {
	if im.Remote == "" {
		return nil, "", nil
	}
	if rcs == nil {
		return nil, "", errors.New("nil RemoteContents")
	}

	ri, ok := rcs.Images[im.Name]
	if !ok {
		return nil, "", errors.Errorf("image %s not found", im.Name)
	}
	targetVersion := im.Reference
	if targetVersion == DefaultTagRef {
		targetVersion = ri.defaultVersion
	}
	for _, vers := range ri.versions {
		if vers.version == targetVersion {
			if vers.location == "" {
				return nil, "", errEmptyLocation
			}
			path := vers.location
			if !strings.Contains(path, "://") {
				path = "./" + path
			}
			location, err := url.Parse(path)
			if err != nil {
				return nil, "", err
			}
			return location, vers.hash, nil
		}
	}

	return nil, "", errors.Errorf("image %s:%s not found", im.Name, im.Reference)
}

// FetchImage checks and fetch an image archive if available on a known remote.
func (rc *RemotesCache) FetchImage(ctx context.Context, im Image, versionedStorePath string) error {
	if rc == nil {
		return errNilRemotesCache
	}
	baseURL, location, hash, err := rc.CheckAvailable(im)
	if err != nil {
		return err
	}
	if baseURL == nil || location == nil {
		return nil
	}

	switch baseURL.Scheme {
	case "file":
		return nil
	case "https", "http":
		tries := 0
		for {
			tries++
			err := rc.downloadArchive(ctx, baseURL, location, versionedStorePath, hash)
			ctxErr := ctx.Err()
			if err == nil && ctxErr == nil {
				return nil
			}
			if ctxErr != nil {
				return ctxErr
			}
			logrus.WithFields(logrus.Fields{
				"attempt":   tries,
				"name":      im.Name,
				"reference": im.Reference,
				"remote":    im.Remote,
				"error":     err,
			}).Error("failed to fetch")
			time.Sleep(8 * time.Second)
		}
	default:
		return errors.Errorf("unsupported scheme while trying to fetch %s", baseURL.String())
	}
}

// downloadArchive downloads an image archive from a remote.
func (rc *RemotesCache) downloadArchive(ctx context.Context, baseURL *url.URL, location *url.URL, baseDir string, hash string) error {
	fileName := path.Base(location.String())
	if !strings.HasSuffix(fileName, ".torcx.tgz") && !strings.HasSuffix(fileName, ".torcx.squashfs") {
		return errors.Errorf("invalid extension for image archive %s", fileName)
	}
	targetPath := filepath.Join(baseDir, fileName)
	tmpFile, err := ioutil.TempFile(baseDir, ".fetchimg")
	if err != nil {
		return err
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)
	defer tmpFile.Close()
	bufwr := bufio.NewWriter(tmpFile)
	defer bufwr.Flush()

	fullURL := baseURL.ResolveReference(location)
	logrus.WithFields(logrus.Fields{
		"url": fullURL.String(),
	}).Info("downloading image archive from remote")
	req, err := http.NewRequest("GET", fullURL.String(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buf := make([]byte, 32*1024)
	if err := ctxcopy.Copy(ctx, bufwr, resp.Body, buf); err != nil {
		return err
	}
	if err := bufwr.Flush(); err != nil {
		return errors.Wrapf(err, "failed to flush %s", tmpName)
	}
	if err := tmpFile.Close(); err != nil {
		return errors.Wrapf(err, "failed to close %s", tmpName)
	}
	if err := os.Chmod(tmpName, 0755); err != nil {
		return errors.Wrapf(err, "failed to chmod %s", tmpName)
	}

	if hash != "" {
		valid, err := validateHash(tmpName, hash)
		if err != nil {
			return errors.Wrapf(err, "failed to validate %s", targetPath)
		}
		if !valid {
			return errors.Errorf("mismatching hash for %s", targetPath)
		}
	}
	if err := os.Rename(tmpName, targetPath); err != nil {
		return errors.Wrapf(err, "failed to save %s", targetPath)
	}

	logrus.WithFields(logrus.Fields{
		"path": targetPath,
	}).Debug("image fetched")
	return nil
}

func validateHash(path string, hash string) (bool, error) {
	fp, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer fp.Close()
	d, err := digest.Parse(strings.Replace(hash, "-", ":", 1))
	if err != nil {
		return false, errors.Wrap(err, "could not understand package hash")
	}

	verifier := d.Verifier()
	if _, err := io.Copy(verifier, bufio.NewReader(fp)); err != nil {
		return false, errors.Wrap(err, "could not read file for hash validation")
	}

	return verifier.Verified(), nil
}
