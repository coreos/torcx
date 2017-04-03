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
	"os"
	"path/filepath"

	"bufio"
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// CurrentProfileName returns the name of the currently running profile
func CurrentProfileName() (string, error) {
	var profile string

	meta, err := ReadMetadata(FUSE_PATH)
	if err != nil {
		return "", err
	}

	profile, ok := meta[FUSE_PROFILE_NAME]
	if !ok {
		return "", errors.New("unable to determine current profile name")
	}

	if profile == "" {
		return "", errors.New("invalid profile name")
	}

	return profile, nil
}

// CurrentProfilePath returns the path of the currently running profile
func CurrentProfilePath() (string, error) {
	var path string

	meta, err := ReadMetadata(FUSE_PATH)
	if err != nil {
		return "", err
	}

	path, ok := meta[FUSE_PROFILE_PATH]
	if !ok {
		return "", errors.New("unable to determine current profile path")
	}

	if path == "" {
		return "", errors.New("invalid profile path")
	}

	return path, nil
}

// ReadCurrentProfile returns the content of the currently running profile
func ReadCurrentProfile() (Images, error) {
	path, err := CurrentProfilePath()
	if err != nil {
		return Images{}, err
	}

	return ReadProfile(path)
}

// ReadProfile returns the content of a specific profile
func ReadProfile(path string) (Images, error) {
	fp, err := os.Open(path)
	if err != nil {
		return Images{}, err
	}
	defer fp.Close()

	var manifest ProfileManifestV0
	jsonIn := json.NewDecoder(bufio.NewReader(fp))
	err = jsonIn.Decode(&manifest)
	if err != nil {
		return Images{}, err
	}

	// TODO(lucab): perform semantic validation

	return manifest.Value, nil
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

		if inInfo.IsDir() && name != "profiles.d" {
			return filepath.SkipDir
		}

		if inInfo.Mode().IsRegular() {
			if parentDir != "profiles.d" {
				return filepath.SkipDir
			}

			// TODO(lucab): perhaps require .json file suffix?

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
