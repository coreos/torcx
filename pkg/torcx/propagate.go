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
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const ()

// propagateUnits installs all systemd unit and unit-like files as transient
// units in /run/systemd.
// Units are taken from /usr/lib/systemd
func propagateUnits(applyCfg *ApplyConfig, imageRoot string) error {
	srcDir := filepath.Join(imageRoot, "usr", "lib", "systemd")
	st, err := os.Stat(srcDir)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if !st.IsDir() {
		return nil
	}

	_, err = os.Stat(SYSTEMD_DIR)
	if err != nil {
		return errors.Wrapf(err, "error checking for systemd volatile directory %s", SYSTEMD_DIR)
	}

	// Walk the source directory
	err = filepath.Walk(srcDir, func(srcPath string, srcStat os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories - we'll do a mkdirall later
		if srcStat.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, srcPath)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		dstPath := filepath.Join(SYSTEMD_DIR, relPath)
		_, dstStatErr := os.Stat(dstPath)
		if dstStatErr == nil {
			return fmt.Errorf("Cannot copy unit file %s to %s - destination exists", srcPath, dstPath)
		} else if !os.IsNotExist(dstStatErr) {
			return err
		}

		// Copy src to dest
		// * If src is symlink, duplicate link contents exactly
		// * If src is file, symlink
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return errors.Wrapf(err, "failed to create dest dir %s", filepath.Dir(dstPath))
		}

		if srcStat.Mode()&os.ModeSymlink > 0 {
			// Duplicate symlink value exactly here.
			// We're not touching the symlink contents because symlinks are sort
			// of special with systemd. They're most often used to indicate a
			// particular unit as wanted by a target, by making a link.
			// While systemd will follow links to unit files and the like,
			// they're not usually used.
			linkDest, err := os.Readlink(srcPath)
			if err != nil {
				return errors.Wrapf(err, "could not read link %s", srcPath)
			}
			return os.Symlink(linkDest, dstPath)
		} else if srcStat.Mode().IsRegular() {
			// Normal files: symlink in to the systemd volatile directory
			logrus.WithFields(logrus.Fields{
				"src": srcPath,
				"dst": dstPath,
			}).Debug("installing unit file")
			return os.Symlink(srcPath, dstPath)
		} else {
			logrus.Warnf("Skipping systemd non-file %s", srcPath)
		}
		return nil
	})

	return nil
}
