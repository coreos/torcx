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

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
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
	return ioutil.WriteFile(cc.NextProfile(), []byte(name), 0644)
}

// ReadCurrentProfile returns the content of the currently running profile
func ReadCurrentProfile() (Images, error) {
	path, err := CurrentProfilePath()
	if err != nil {
		return Images{}, err
	}

	return ReadProfilePath(path)
}

// ReadProfilePath returns the content of a specific profile, specified via path.
func ReadProfilePath(path string) (Images, error) {
	fp, err := os.Open(path)
	if err != nil {
		return Images{}, err
	}
	defer fp.Close()

	return readProfileReader(bufio.NewReader(fp))
}

// readProfileReader returns the content of a specific profile, specified via a reader.
func readProfileReader(in io.Reader) (Images, error) {
	var manifest ProfileManifestV0

	jsonIn := json.NewDecoder(in)
	err := jsonIn.Decode(&manifest)
	if err == io.EOF {
		return Images{}, nil
	}
	if err != nil {
		return Images{}, err
	}

	// TODO(lucab): perform semantic validation

	return manifest.Value, nil
}

// getProfile reads a profile from the given path, does unmarshal json format,
// to return the interpreted profile manifest.
func getProfile(profilePath string) (ProfileManifestV0, error) {
	var manifest ProfileManifestV0

	b, err := ioutil.ReadFile(profilePath)
	if err != nil {
		return ProfileManifestV0{}, err
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		return ProfileManifestV0{}, err
	}

	return manifest, nil
}

// putProfile does marshal the given profile manifest into json format,
// to write the profile into the given path.
func putProfile(profilePath string, perm os.FileMode, manifest ProfileManifestV0) error {
	b, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(profilePath, b, perm)
}

func AddToProfile(profilePath string, im Image) error {
	st, err := os.Stat(profilePath)
	if err != nil {
		return err
	}

	manifest, err := getProfile(profilePath)
	if err != nil {
		return err
	}

	// Update if existing
	found := false
	for idx, mim := range manifest.Value.Images {
		if mim.Name == im.Name {
			manifest.Value.Images[idx] = im
			found = true
			break
		}
	}

	// Add otherwise
	if !found {
		manifest.Value.Images = append(manifest.Value.Images, im)
	}

	return putProfile(profilePath, st.Mode().Perm(), manifest)
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

func mergeProfiles(applyCfg *ApplyConfig) (Images, error) {
	var mergedImages Images

	if applyCfg == nil {
		return Images{}, errors.New("missing apply configuration")
	}
	localProfiles, err := ListProfiles(applyCfg.ProfileDirs())
	if err != nil {
		return Images{}, errors.Wrap(err, "profiles listing failed")
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
			return Images{}, errors.Errorf("profile %q not found", lp)
		}
		fp, err := os.Open(profilePath)
		if err != nil {
			return Images{}, errors.Wrapf(err, "opening profile %q", profilePath)
		}
		defer fp.Close()
		images, err := readProfileReader(bufio.NewReader(fp))
		if err != nil && err != io.EOF {
			return Images{}, errors.Wrapf(err, "reading profile %q", profilePath)
		}
		mergedImages = mergeImages(mergedImages, images)
	}
	return mergedImages, nil
}

// mergeImages merges two arrays of images ("lower" and "upper"), keeping their relative order.
// Images from "upper" are appended at the end, and can override images from "lower".
// nil references and names are excluded from the final array.
func mergeImages(lower Images, upper Images) Images {
	// TODO(lucab): perhaps trade-off time/memory here with a linked-hashmap
	merged := Images{Images: make([]Image, 0, len(lower.Images)+len(upper.Images))}
	lowerImages := make(map[string]bool, len(lower.Images))
	upperImages := make(map[string]bool, len(upper.Images))

	// Compute the set of images to keep
	for _, image := range lower.Images {
		if image.Reference != "" {
			lowerImages[image.Name] = true
		}
	}
	for _, image := range upper.Images {
		delete(lowerImages, image.Name)
		if image.Reference != "" {
			upperImages[image.Name] = true
		}
	}

	// Merge in order
	for _, image := range lower.Images {
		if image.Name != "" && lowerImages[image.Name] {
			merged.Images = append(merged.Images, image)
		}
	}
	for _, image := range upper.Images {
		if image.Name != "" && upperImages[image.Name] {
			merged.Images = append(merged.Images, image)
		}
	}

	return merged
}
