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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/torcx/pkg/untar"
	"github.com/pkg/errors"
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

	originPath, ok := localProfiles[applyCfg.Profile]
	if !ok {
		return fmt.Errorf("profile %q not found", applyCfg.Profile)
	}

	opp, err := os.Open(originPath)
	if err != nil {
		return err
	}
	defer opp.Close()

	images, err := readProfileReader(bufio.NewReader(opp))
	if err != nil {
		return err
	}
	if len(images.Images) == 0 {
		return nil
	}

	storeCache, err := NewStoreCache(applyCfg.StorePaths)
	if err != nil {
		return err
	}

	for _, im := range images.Images {
		tgzArchive, err := storeCache.ArchiveFor(im)
		if err != nil {
			return err
		}

		err = unpackTgz(applyCfg, tgzArchive.Filepath, im.Name)
		if err != nil {
			return err
		}

		// TODO(lucab): scan/symlink/extract binaries and systemd-units

		logrus.WithFields(logrus.Fields{
			"name":      im.Name,
			"reference": im.Reference,
			"path":      originPath,
		}).Debug("image unpacked")
	}

	runProfilePath := filepath.Join(applyCfg.RunDir, "profile")
	rpp, err := os.Create(runProfilePath)
	if err != nil {
		return err
	}
	defer rpp.Close()

	if n, err := opp.Seek(0, io.SeekStart); err != nil || n != 0 {
		return fmt.Errorf("seek failed")
	}

	_, err = io.Copy(rpp, opp)
	if err != nil {
		return err
	}
	err = os.Chmod(runProfilePath, 0444)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"name":             applyCfg.Profile,
		"original profile": originPath,
		"sealed profile":   runProfilePath,
	}).Debug("profile applied")

	return nil
}

// SealSystemState is a one-time-op which seals the current state of the system,
// after a torcx profile has been applied to it.
func SealSystemState(applyCfg *ApplyConfig) error {
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
		fmt.Sprintf("%s=%q", FUSE_BINDIR, filepath.Join(applyCfg.RunDir, "bin/")),
		fmt.Sprintf("%s=%q", FUSE_UNPACKDIR, filepath.Join(applyCfg.RunDir, "unpack/")),
	}

	for _, line := range content {
		_, err = fp.WriteString(line + "\n")
		if err != nil {
			return errors.Wrap(err, "writing seal content")
		}
	}

	logrus.WithFields(logrus.Fields{
		"path":    FUSE_PATH,
		"content": content,
	}).Debug("system state sealed")

	return nil
}

func ensurePaths(applyCfg *ApplyConfig) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

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

func unpackTgz(applyCfg *ApplyConfig, tgzPath, imageName string) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	if tgzPath == "" || imageName == "" {
		return errors.New("missing unpack source")
	}

	topDir := filepath.Join(applyCfg.RunDir, "unpack", imageName)
	if _, err := os.Stat(topDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(topDir, 0755); err != nil {
			return err
		}
	}

	fp, err := os.Open(tgzPath)
	if err != nil {
		return errors.Wrapf(err, "opening %q", tgzPath)
	}
	defer fp.Close()

	gr, err := gzip.NewReader(fp)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	untarCfg := untar.ExtractCfg{}.Default()
	err = untar.ChrootUntar(tr, topDir, untarCfg)
	if err != nil {
		return errors.Wrapf(err, "unpacking %q", tgzPath)
	}

	return nil
}
