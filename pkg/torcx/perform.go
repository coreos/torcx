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
	"syscall"

	"github.com/Sirupsen/logrus"
	pkgtar "github.com/coreos/torcx/pkg/tar"
	"github.com/pkg/errors"
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

	err = setupPaths(applyCfg)
	if err != nil {
		return errors.Wrap(err, "profile setup")
	}

	localProfiles, err := ListProfiles(applyCfg.ProfileDirs())
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

		// phase 1: unpack image
		imageRoot, err := unpackTgz(applyCfg, tgzArchive.Filepath, im.Name)
		if err != nil {
			return err
		}

		// phase 2: propagate assets
		bins, err := propagateBins(applyCfg, imageRoot)
		if err != nil {
			return err
		}

		// And symlink in systemd unit files
		err = propagateUnits(applyCfg, imageRoot)
		if err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"name":              im.Name,
			"reference":         im.Reference,
			"path":              imageRoot,
			"provided binaries": bins,
		}).Debug("image unpacked")
	}

	rpp, err := os.Create(applyCfg.RunProfile())
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
	err = os.Chmod(applyCfg.RunProfile(), 0444)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"name":             applyCfg.Profile,
		"original profile": originPath,
		"sealed profile":   applyCfg.RunProfile(),
	}).Debug("profile applied")

	// phase 3: seal system state
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
		fmt.Sprintf("%s=%q", FUSE_PROFILE_PATH, applyCfg.RunProfile()),
		fmt.Sprintf("%s=%q", FUSE_BINDIR, applyCfg.RunBinDir()),
		fmt.Sprintf("%s=%q", FUSE_UNPACKDIR, applyCfg.RunUnpackDir()),
	}

	for _, line := range content {
		_, err = fp.WriteString(line + "\n")
		if err != nil {
			return errors.Wrap(err, "writing seal content")
		}
	}

	// Remount the unpackdir RO
	if err := syscall.Mount(applyCfg.RunUnpackDir(), applyCfg.RunUnpackDir(),
		"", syscall.MS_REMOUNT|syscall.MS_RDONLY, ""); err != nil {

		return errors.Wrap(err, "failed to remount read-only")
	}

	logrus.WithFields(logrus.Fields{
		"path":    FUSE_PATH,
		"content": content,
	}).Debug("system state sealed")

	return nil
}

func setupPaths(applyCfg *ApplyConfig) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}

	paths := []string{applyCfg.BaseDir, applyCfg.RunDir, applyCfg.ConfDir}
	paths = append(paths, applyCfg.RunBinDir())
	paths = append(paths, applyCfg.RunUnpackDir())
	paths = append(paths, applyCfg.UserProfileDir())
	// TODO: implement auth...
	//paths = append(paths, cc.AuthDir())

	for _, d := range paths {
		if _, err := os.Stat(d); err != nil && os.IsNotExist(err) {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
		}
	}

	logrus.WithField("dest", applyCfg.RunUnpackDir()).Debug("mounting tmpfs to unpack directory")

	// Now, mount a tmpfs directory to the unpack directory
	// We need to do this because, unsurprisingly, "/run" is noexec
	if err := syscall.Mount("none", applyCfg.RunUnpackDir(), "tmpfs", 0, ""); err != nil {
		return errors.Wrap(err, "Failed to mount unpack dir")
	}

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
