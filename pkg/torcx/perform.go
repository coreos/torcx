// Copyright 2017 CoreOS Inc.
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
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	pkgtar "github.com/coreos/torcx/pkg/tar"
	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

var (
	// ErrUnknownOSVersionID is the error returned on generic os-release parsing errors
	ErrUnknownOSVersionID = errors.New(`unable to parse "VERSION_ID" from os-release`)
)

// ApplyProfile is called at boot-time to apply the configured profile
// system-wide. Apply operation is split in three phases:
//  * unpack: all images are unpacked to their own dedicated path under UnpackDir
//  * propagate: executable assets are propagated into the system;
//    this includes symlinking binaries into BinDir and installing systemd
//    transient units.
//  * seal: system state is frozen, profile and metadata written to RunDir
func ApplyProfile(applyCfg *ApplyConfig) error {
	var err error
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	err = cleanPaths(applyCfg)
	if err != nil {
		return errors.Wrap(err, "clean paths")
	}
	err = setupPaths(applyCfg)
	if err != nil {
		return errors.Wrap(err, "profile setup")
	}

	images, err := mergeProfiles(applyCfg)
	if err != nil {
		return err
	}
	if len(images.Images) > 0 {
		if err := applyImages(applyCfg, images); err != nil {
			return err
		}
	}

	runProfile := ProfileManifestV0{
		Kind:  ProfileManifestV0K,
		Value: images,
	}
	rpp, err := os.Create(applyCfg.RunProfile())
	if err != nil {
		return err
	}
	defer rpp.Close()
	bufwr := bufio.NewWriter(rpp)
	defer bufwr.Flush()
	enc := json.NewEncoder(bufwr)
	enc.SetIndent("", "  ")
	if err := enc.Encode(runProfile); err != nil {
		return errors.Wrapf(err, "writing %q", applyCfg.RunProfile())
	}

	err = os.Chmod(applyCfg.RunProfile(), 0444)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"upper profile":  applyCfg.UpperProfile,
		"sealed profile": applyCfg.RunProfile(),
	}).Debug("profile applied")
	return nil
}

// applyImages unpacks and propagates assets from a list of images.
func applyImages(applyCfg *ApplyConfig, images Images) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	storeCache, err := NewStoreCache(applyCfg.StorePaths)
	if err != nil {
		return err
	}

	// Unpack all images, continuing on error
	failedImages := []Image{}

	for _, im := range images.Images {
		// Some log fields we keep using
		logFields := logrus.Fields{
			"image":     im.Name,
			"reference": im.Reference,
		}

		tgzArchive, err := storeCache.ArchiveFor(im)
		if err != nil {
			logrus.WithFields(logFields).Error(err)
			failedImages = append(failedImages, im)
			continue
		}

		imageRoot, err := unpackTgz(applyCfg, tgzArchive.Filepath, im.Name)
		if err != nil {
			failedImages = append(failedImages, im)
			logrus.WithFields(logFields).Error("failed to unpack: ", err)
			continue
		}
		logFields["path"] = imageRoot
		logrus.WithFields(logFields).Debug("image unpacked")

		assets, err := retrieveAssets(applyCfg, imageRoot)
		if err != nil {
			failedImages = append(failedImages, im)
			logrus.WithFields(logFields).Error("failed retrieving assets from image: ", err)
			continue
		}

		if len(assets.Binaries) > 0 {
			if err := propagateBins(applyCfg, imageRoot, assets.Binaries); err != nil {
				failedImages = append(failedImages, im)
				logrus.WithFields(logFields).WithField("assets", assets.Binaries).Error("failed to propagate binaries: ", err)
				continue
			}
			logrus.WithFields(logFields).WithField("assets", assets.Binaries).Debug("binaries propagated")
		}

		if len(assets.Network) > 0 {
			if err := propagateNetworkdUnits(applyCfg, imageRoot, assets.Network); err != nil {
				failedImages = append(failedImages, im)
				logrus.WithFields(logFields).WithField("assets", assets.Network).Error("failed to propagate networkd units: ", err)
				continue
			}

			logrus.WithFields(logFields).WithField("assets", assets.Network).Debug("networkd units propagated")
		}

		if len(assets.Units) > 0 {
			if err := propagateSystemdUnits(applyCfg, imageRoot, assets.Units); err != nil {
				failedImages = append(failedImages, im)
				logrus.WithFields(logFields).WithField("assets", assets.Units).Error("failed to propagate systemd units: ", err)
				continue
			}
			logrus.WithFields(logFields).WithField("assets", assets.Units).Debug("systemd units propagated")
		}

		if len(assets.Sysusers) > 0 {
			if err := propagateSysusersUnits(applyCfg, imageRoot, assets.Sysusers); err != nil {
				failedImages = append(failedImages, im)
				logrus.WithFields(logFields).WithField("assets", assets.Sysusers).Error("failed to propagate sysusers: ", err)
				continue
			}
			logrus.WithFields(logFields).WithField("assets", assets.Sysusers).Debug("sysusers propagated")
		}

		if len(assets.Tmpfiles) > 0 {
			if err := propagateTmpfilesUnits(applyCfg, imageRoot, assets.Tmpfiles); err != nil {
				failedImages = append(failedImages, im)
				logrus.WithFields(logFields).WithField("assets", assets.Units).Error("failed to propagate tmpfiles: ", err)
				continue
			}
			logrus.WithFields(logFields).WithField("assets", assets.Units).Debug("tmpfiles propagated")
		}

		// TODO(lucab): evaluate and propagate more units types
	}

	if len(failedImages) > 0 {
		return fmt.Errorf("failed to install %d images", len(failedImages))
	}

	return nil
}

// SealSystemState is a one-time-op which seals the current state of the system,
// after a torcx profile has been applied to it.
func SealSystemState(applyCfg *ApplyConfig) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	dirname := filepath.Dir(SealPath)
	if _, err := os.Stat(dirname); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dirname, 0755); err != nil {
			return err
		}
	}

	fp, err := os.Create(SealPath)
	if err != nil {
		return err
	}
	defer fp.Close()

	content := []string{
		fmt.Sprintf("%s=%q", SealLowerProfiles, strings.Join(applyCfg.LowerProfiles, ":")),
		fmt.Sprintf("%s=%q", SealUpperProfile, applyCfg.UpperProfile),
		fmt.Sprintf("%s=%q", SealRunProfilePath, applyCfg.RunProfile()),
		fmt.Sprintf("%s=%q", SealBindir, applyCfg.RunBinDir()),
		fmt.Sprintf("%s=%q", SealUnpackdir, applyCfg.RunUnpackDir()),
	}

	for _, line := range content {
		_, err = fp.WriteString(line + "\n")
		if err != nil {
			return errors.Wrap(err, "writing seal content")
		}
	}

	// Remount the unpackdir RO
	if err := unix.Mount(applyCfg.UnpackDir(), applyCfg.UnpackDir(),
		"", unix.MS_REMOUNT|unix.MS_RDONLY, ""); err != nil {

		return errors.Wrap(err, "failed to remount read-only")
	}

	logrus.WithFields(logrus.Fields{
		"path":    SealPath,
		"content": content,
	}).Debug("system state sealed")

	return nil
}

func cleanPaths(applyCfg *ApplyConfig) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	// TODO: we can safely leave the persistent unpackdir if we have a good way
	// to uniquely identify it (e.g. the hash of its contents or such).
	// For now, deal with extra disk churn; delete and recreate it even when not
	// necessary.

	if err := os.RemoveAll(applyCfg.UnpackDir()); err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unable to remove previous unpackdir")
	}
	return nil
}

func setupPaths(applyCfg *ApplyConfig) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	paths := []string{
		// RunDir is the first path created, signaling that torcx run
		applyCfg.RunDir,
		applyCfg.BaseDir,
		applyCfg.ConfDir,
		applyCfg.RunBinDir(),
		applyCfg.UnpackDir(),
		applyCfg.RunUnpackDir(),
		applyCfg.UserProfileDir(),
	}

	for _, d := range paths {
		if _, err := os.Stat(d); err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		}
	}

	// Bind the UnpackDir over to the RunUnpackDir.
	// This is done to provide a transient location for currently active images
	// (the RunUnpackDir, typically a tmpfs which will vanish on reboot), while
	// not actually storing data in the tmpfs (memory usage).
	// This decoupling means that, in theory, multiple UnpackDirs could exist and
	// then, at boot-time, be selected between by choosing which to bindmount.
	// In addition, this is done for backwards compatibility; previously the
	// 'UnpackDir' did not exist and the 'RunUnpackDir' was both the source of
	// truth and store of data.
	if err := unix.Mount(applyCfg.UnpackDir(), applyCfg.RunUnpackDir(), "", unix.MS_BIND|unix.MS_REC|unix.MS_SLAVE, ""); err != nil {
		return errors.Wrap(err, "failed to bindmount unpackdir")
	}

	logrus.WithField("target", applyCfg.RunUnpackDir()).Debug("mounted tmpfs")
	return nil
}

// unpackTgz renders a tgz rootfs, returning the target top directory.
func unpackTgz(applyCfg *ApplyConfig, tgzPath, imageName string) (string, error) {
	if applyCfg == nil {
		return "", errors.New("missing apply configuration")
	}

	if tgzPath == "" || imageName == "" {
		return "", errors.New("missing unpack source")
	}

	topDir := filepath.Join(applyCfg.RunUnpackDir(), imageName)
	if _, err := os.Stat(topDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(topDir, 0755); err != nil {
			return "", err
		}
	}

	fp, err := os.Open(tgzPath)
	if err != nil {
		return "", errors.Wrapf(err, "opening %q", tgzPath)
	}
	defer fp.Close()

	gr, err := gzip.NewReader(fp)
	if err != nil {
		return "", err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	untarCfg := pkgtar.ExtractCfg{}.Default()
	err = pkgtar.ChrootUntar(tr, topDir, untarCfg)
	if err != nil {
		return "", errors.Wrapf(err, "unpacking %q", tgzPath)
	}

	return topDir, nil
}

// CurrentOsVersionID parses an os-release file to extract the VERSION_ID.
//
// For more details about the expect format of the os-release file, see
// https://www.freedesktop.org/software/systemd/man/os-release.html
func CurrentOsVersionID(path string) (string, error) {
	if path == "" {
		path = OsReleasePath
	}
	fp, err := os.Open(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open %q", path)
	}
	defer fp.Close()
	return parseOsVersionID(fp)
}

// parseOsVersionID is the parser for os-release.
func parseOsVersionID(rd io.Reader) (string, error) {
	ver := ""

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
		if parts[0] == "VERSION_ID" {
			ver = parts[1]
			break
		}
	}
	if sc.Err() != nil {
		return "", errors.Wrap(sc.Err(), "failed to parse os-release file")
	}
	if ver == "" {
		return "", ErrUnknownOSVersionID
	}
	return ver, nil
}
