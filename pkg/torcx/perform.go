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
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"path/filepath"
)

// ApplyProfile is called at boot-time to apply the configured profile
// system-wide
func ApplyProfile(applyCfg *ApplyConfig) error {
	var err error
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	err = ensurePaths(applyCfg)
	if err != nil {
		return errors.Wrap(err, "profile setup")
	}

	profileDirs := []string{
		filepath.Join(VENDOR_DIR, "profiles.d"),
		filepath.Join(applyCfg.ConfDir, "profiles.d"),
	}
	localProfiles, err := ListProfiles(profileDirs)
	if err != nil {
		return errors.Wrap(err, "profiles listing failed")
	}

	path, ok := localProfiles[applyCfg.Profile]
	if !ok {
		return fmt.Errorf("profile %q not found", applyCfg.Profile)
	}

	bundles, err := ReadProfile(path)
	if err != nil {
		return err
	}
	if len(bundles.Archives) == 0 {
		return nil
	}

	storeCache, err := NewStoreCache(applyCfg.StorePaths)
	if err != nil {
		return err
	}

	for _, pkg := range bundles.Archives {
		path, err := storeCache.LookupReference(pkg)
		if err != nil {
			return err
		}

		// TODO(lucab): render bundle refs

		logrus.WithFields(logrus.Fields{
			"bundle name": pkg.Image,
			"reference":   pkg.Reference,
			"path":        path,
		}).Debug("bundle/reference unpacked")
	}

	logrus.WithFields(logrus.Fields{
		"fuse path": FUSE_PATH,
		"profile":   applyCfg.Profile,
	}).Debug("profile applied")

	return nil
}

// BlowFuse blows the system-wide torcx fuse
func BlowFuse(applyCfg *ApplyConfig) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	dirname := filepath.Dir(FUSE_PATH)
	if _, err := os.Stat(dirname); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dirname, 0755); err != nil {
			return err
		}
	}

	fp, err := os.Create(FUSE_PATH)
	if err != nil {
		return err
	}
	defer fp.Close()

	content := []string{
		fmt.Sprintf("%s=%q", FUSE_PROFILE_NAME, applyCfg.Profile),
		fmt.Sprintf("%s=%q", FUSE_PROFILE_PATH, filepath.Join(applyCfg.RunDir, "profile")),
		fmt.Sprintf("%s=%q", FUSE_BINDIR, filepath.Join(applyCfg.RunDir, "bin")),
		fmt.Sprintf("%s=%q", FUSE_UNPACKDIR, filepath.Join(applyCfg.RunDir, "unpack")),
	}

	for _, line := range content {
		_, err = fp.WriteString(line + "\n")
		if err != nil {
			return errors.Wrap(err, "writing fuse content")
		}
	}

	logrus.WithFields(logrus.Fields{
		"path":    FUSE_PATH,
		"content": content,
	}).Debug("fuse blown")

	return nil
}

func ensurePaths(applyCfg *ApplyConfig) error {
	paths := []string{applyCfg.BaseDir, applyCfg.RunDir, applyCfg.ConfDir}
	// TODO(lucab): move derived dirs to getters
	paths = append(paths, filepath.Join(applyCfg.RunDir, "bin"))
	paths = append(paths, filepath.Join(applyCfg.RunDir, "unpack"))
	paths = append(paths, filepath.Join(applyCfg.ConfDir, "auth.d"))
	paths = append(paths, filepath.Join(applyCfg.ConfDir, "profiles.d"))

	for _, d := range paths {
		if _, err := os.Stat(d); err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		}
	}

	return nil
}
