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
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// DefaultLowerProfiles are the default lower profiles (for vendor and oem entries)
var DefaultLowerProfiles = []string{VendorProfileName, OemProfileName}

// CurrentProfileNames returns the name of the currently running user and vendor profiles
func CurrentProfileNames() (string, []string, error) {
	meta, err := ReadMetadata(SealPath)
	if err != nil {
		return "", nil, err
	}
	upperProfile, ok := meta[SealUpperProfile]
	if !ok {
		return "", nil, errors.New("unable to determine current upper profile name")
	}
	lowerString, ok := meta[SealLowerProfiles]
	if !ok {
		return "", nil, errors.New("unable to determine current lower profile names")
	}
	lowerProfiles := strings.Split(lowerString, ":")

	return upperProfile, lowerProfiles, nil
}

// CurrentProfilePath returns the path of the currently running profile
func CurrentProfilePath() (string, error) {
	var path string

	meta, err := ReadMetadata(SealPath)
	if err != nil {
		return "", err
	}

	path, ok := meta[SealRunProfilePath]
	if !ok {
		return "", errors.New("unable to determine current profile path")
	}

	if path == "" {
		return "", errors.New("invalid profile path")
	}

	return path, nil
}

// NextProfileName determines which profile will be used for the next apply.
func (cc *CommonConfig) NextProfileName() (string, error) {
	fc, err := ioutil.ReadFile(cc.NextProfile())
	if err != nil {
		return "", errors.Wrapf(err, "unable to read profile file")
	}

	profileName := strings.TrimSpace(string(fc))
	profileName = strings.TrimSuffix(profileName, ".json")

	// Check that the profile exists
	profiles, err := ListProfiles(cc.ProfileDirs())
	if err != nil {
		return "", errors.Wrap(err, "could not list profiles")
	}
	if profileName == "" {
		return "", errors.New("missing profile name")
	}
	if _, ok := profiles[profileName]; !ok {
		return "", errors.Errorf("profile %q not found", profileName)
	}

	return profileName, nil
}

// SetNextProfileName writes the given profile name as active for the next boot.
func (cc *CommonConfig) SetNextProfileName(name string) error {
	if err := os.MkdirAll(cc.UserProfileDir(), 0755); err != nil {
		return err
	}
	line := strings.TrimSpace(name) + "\n"
	return ioutil.WriteFile(cc.NextProfile(), []byte(line), 0644)
}

// ReadCurrentProfile returns the content of the currently running profile
func ReadCurrentProfile() ([]Image, error) {
	path, err := CurrentProfilePath()
	if err != nil {
		return nil, err
	}

	return ReadProfilePath(path)
}

// ReadProfilePath returns the content of a specific profile, specified via path.
func ReadProfilePath(path string) ([]Image, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	return readProfileReader(bufio.NewReader(fp))
}

// readProfileReader returns the content of a specific profile, specified via a reader.
func readProfileReader(in io.Reader) ([]Image, error) {
	var container kindValueJSON
	err := json.NewDecoder(in).Decode(&container)
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	switch container.Kind {
	case ProfileManifestV0K:
		manifest := ProfileManifestV0JSON{
			Kind: container.Kind,
		}
		if err := json.Unmarshal(container.Value, &manifest.Value); err != nil {
			return nil, err
		}
		return ImagesFromJSONV0(manifest.Value), nil
	}

	return nil, errors.Errorf("unknown profile kind %s", container.Kind)
}

// AddToProfile adds an image to an existing profile.
func AddToProfile(profilePath string, im Image) error {
	st, err := os.Stat(profilePath)
	if err != nil {
		return err
	}

	if v0Profile, err := getProfileV0(profilePath); err == nil {
		return addToProfileV0(profilePath, st.Mode().Perm(), v0Profile, &im)
	}

	return errors.Wrapf(err, "unable to unmarshal profile to %s", profilePath)
}

// getProfileV0 reads a profile from the given path, does unmarshal json format,
// to return the interpreted profile manifest.
func getProfileV0(profilePath string) (ProfileManifestV0JSON, error) {
	var manifest ProfileManifestV0JSON
	empty := ProfileManifestV0JSON{
		Kind: ProfileManifestV0K,
	}

	b, err := ioutil.ReadFile(profilePath)
	if err != nil {
		if err == io.EOF {
			return empty, nil
		}
		return ProfileManifestV0JSON{}, err
	}
	if len(b) == 0 {
		return empty, nil
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		return ProfileManifestV0JSON{}, err
	}
	if manifest.Kind != ProfileManifestV0K {
		return manifest, errors.Errorf("expected manifest kind %s, got %s", ProfileManifestV0K, manifest.Kind)
	}

	return manifest, nil
}

// addToProfileV0 does marshal the given profile manifest into json format,
// to write the profile into the given path.
func addToProfileV0(profilePath string, perm os.FileMode, manifest ProfileManifestV0JSON, im *Image) error {
	// Update if existing
	found := false
	if im == nil {
		found = true
	}
	for idx, mim := range manifest.Value.Images {
		if found {
			break
		}
		if mim.Name == im.Name {
			manifest.Value.Images[idx] = im.ToJSONV0()
			found = true
		}
	}
	// Add otherwise
	if !found {
		manifest.Value.Images = append(manifest.Value.Images, im.ToJSONV0())
	}

	b, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(profilePath, b, perm)
}

// ListProfiles returns a list of all available profiles
func ListProfiles(profileDirs []string) (map[string]string, error) {
	profiles := map[string]string{}

	walkFn := func(inPath string, inInfo os.FileInfo, inErr error) error {
		if inErr != nil {
			return nil
		}
		path := filepath.Clean(inPath)
		name := filepath.Base(path)
		parentDir := filepath.Base(filepath.Dir(path))

		if inInfo.IsDir() && name != "profiles" {
			return filepath.SkipDir
		}

		if !strings.HasSuffix(name, ".json") {
			return nil
		}
		name = strings.TrimSuffix(name, ".json")

		if inInfo.Mode().IsRegular() {
			if parentDir != "profiles" {
				return filepath.SkipDir
			}

			profiles[name] = path
			logrus.WithFields(logrus.Fields{
				"name": name,
				"path": path,
			}).Debug("profile found")
		}

		return nil
	}

	for _, root := range profileDirs {
		if err := filepath.Walk(root, walkFn); err != nil {
			return profiles, err
		}
	}

	return profiles, nil
}

func mergeProfiles(applyCfg *ApplyConfig) ([]Image, error) {
	var mergedImages []Image

	if applyCfg == nil {
		return nil, errors.New("missing apply configuration")
	}
	localProfiles, err := ListProfiles(applyCfg.ProfileDirs())
	if err != nil {
		return nil, errors.Wrap(err, "profiles listing failed")
	}

	// We first filter out non-existing lower profiles
	resProfiles := []string{}
	for _, lowerProfile := range applyCfg.LowerProfiles {
		profilePath, ok := localProfiles[lowerProfile]
		if ok && profilePath != "" {
			resProfiles = append(resProfiles, lowerProfile)
		}

	}
	// Then we check whether we should append an upper profile
	if applyCfg.UpperProfile != "" {
		resProfiles = append(resProfiles, applyCfg.UpperProfile)
	}

	// Then we do a stable merge of images from all profiles (in-order)
	for _, lp := range resProfiles {
		profilePath, ok := localProfiles[lp]
		if !ok || profilePath == "" {
			return nil, errors.Errorf("profile %q not found", lp)
		}
		fp, err := os.Open(profilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "opening profile %q", profilePath)
		}
		defer fp.Close()
		images, err := readProfileReader(bufio.NewReader(fp))
		if err != nil && err != io.EOF {
			return nil, errors.Wrapf(err, "reading profile %q", profilePath)
		}
		mergedImages = mergeImages(mergedImages, images)
	}
	return mergedImages, nil
}

// mergeImages merges two arrays of images ("lower" and "upper"), keeping their relative order.
// Images from "upper" are appended at the end, and can override images from "lower".
// nil references and names are excluded from the final array.
func mergeImages(lower []Image, upper []Image) []Image {
	// TODO(lucab): perhaps trade-off time/memory here with a linked-hashmap
	merged := make([]Image, 0, len(lower)+len(upper))
	lowerImages := make(map[string]bool, len(lower))
	upperImages := make(map[string]bool, len(upper))

	// Compute the set of images to keep
	for _, image := range lower {
		if image.Reference != "" {
			lowerImages[image.Name] = true
		}
	}
	for _, image := range upper {
		delete(lowerImages, image.Name)
		if image.Reference != "" {
			upperImages[image.Name] = true
		}
	}

	// Merge in order
	for _, image := range lower {
		if image.Name != "" && lowerImages[image.Name] {
			merged = append(merged, image)
		}
	}
	for _, image := range upper {
		if image.Name != "" && upperImages[image.Name] {
			merged = append(merged, image)
		}
	}

	return merged
}
