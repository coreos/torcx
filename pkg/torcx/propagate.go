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
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	// manifestPath is the well-known location for image manifest
	manifestPath = "/.torcx/manifest.json"
	// systemdDir is the runtime systemd base path
	// TODO(lucab): possibly not constant, group all link-time parameter together
	systemdDir  = "/run/systemd"
	sysUsersDir = "/run/sysusers.d"
	tmpFilesDir = "/run/tmpfiles.d"
)

func retrieveAssets(applyCfg *ApplyConfig, imageRoot string) (*Assets, error) {
	if applyCfg == nil {
		return nil, errors.New("missing apply configuration")
	}
	if imageRoot == "" {
		return nil, errors.New("missing image top directory")
	}
	assets := &Assets{}
	path := filepath.Join(imageRoot, manifestPath)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Corner-case: missing manifest, no assets to propagate
			return assets, nil
		}
		return nil, err
	}

	var manifest ImageManifestV0
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &manifest); err != nil {
		return nil, err
	}

	return &manifest.Value, nil
}

// propagateNetworkdUnits installs networkd unit files as runtime units (in /run/systemd/network/).
func propagateNetworkdUnits(applyCfg *ApplyConfig, imageRoot string, units []string) error {
	ndUnitsDir := filepath.Join(systemdDir, "network")
	return propagateUnits(applyCfg, imageRoot, units, ndUnitsDir)
}

// propagateSystemdUnits installs systemd unit files as runtime units (in /run/systemd/system/).
func propagateSystemdUnits(applyCfg *ApplyConfig, imageRoot string, units []string) error {
	sdUnitsDir := filepath.Join(systemdDir, "system")
	return propagateUnits(applyCfg, imageRoot, units, sdUnitsDir)
}

// propagateSysusersUnits installs sysusers config in /run/sysusers.d
func propagateSysusersUnits(applyCfg *ApplyConfig, imageRoot string, units []string) error {
	return propagateUnits(applyCfg, imageRoot, units, sysUsersDir)
}

func propagateTmpfilesUnits(applyCfg *ApplyConfig, imageRoot string, units []string) error {
	return propagateUnits(applyCfg, imageRoot, units, tmpFilesDir)
}

// propagateUnits installs unit assets as runtime units for systemd/networkd/etc.
func propagateUnits(applyCfg *ApplyConfig, imageRoot string, units []string, unitsDir string) error {
	if len(units) <= 0 {
		// Corner-case: no units to propagate
		return nil
	}
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}
	if imageRoot == "" {
		return errors.New("missing image top directory")
	}
	if unitsDir == "" {
		return errors.New("missing image target directory")
	}

	if _, err := os.Stat(unitsDir); err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "error checking runtime directory %s", unitsDir)
		}
		if err := os.MkdirAll(unitsDir, 0755); err != nil {
			return errors.Wrapf(err, "error creating runtime directory %s", unitsDir)
		}
	}

	for _, servEntry := range units {
		if servEntry == "" {
			continue
		}
		path := filepath.Join(imageRoot, servEntry)
		if err := symlinkUnitAsset(applyCfg, unitsDir, path); err != nil {
			return err
		}
	}
	return nil
}

// propagateBins symlinks binaries from unpacked image to torcx bindir.
func propagateBins(applyCfg *ApplyConfig, imageRoot string, binaries []string) error {
	if len(binaries) <= 0 {
		// Corner-case: no binaries to propagate
		return nil
	}
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}
	if imageRoot == "" {
		return errors.New("missing image top directory")
	}

	for _, binEntry := range binaries {
		if binEntry == "" {
			continue
		}
		path := filepath.Join(imageRoot, binEntry)
		if err := symlinkBinAsset(applyCfg, applyCfg.RunBinDir(), path); err != nil {
			return err
		}
	}
	return nil
}

// flattenBinAssets propagates a single binary or a directory of binaries,
// flattening all intermediate directories.
func symlinkBinAsset(applyCfg *ApplyConfig, binDir string, asset string) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}
	if asset == "" {
		return errors.New("missing asset path")
	}
	if binDir == "" {
		return errors.New("missing torcx binary directory")
	}

	walkFn := func(inPath string, inInfo os.FileInfo, inErr error) error {
		if inErr != nil {
			return nil
		}
		path := filepath.Clean(inPath)
		baseName := filepath.Base(path)
		newName := filepath.Join(binDir, baseName)

		if inInfo.Mode().IsRegular() || inInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			if _, err := os.Stat(newName); err == nil {
				// Do not overwrite previous assets
				return nil
			}
			return os.Symlink(path, newName)
		}
		return nil
	}

	return filepath.Walk(asset, walkFn)
}

// symlinkUnitAsset propagates a single unit or a directory of units,
// flattening all but the last intermediate directories.
func symlinkUnitAsset(applyCfg *ApplyConfig, unitsDir string, asset string) error {
	if applyCfg == nil {
		return errors.New("missing apply configuration")
	}
	if asset == "" {
		return errors.New("missing asset path")
	}
	if unitsDir == "" {
		return errors.New("missing host units directory")
	}

	// If asset is a directory, keep everything below it unflattened
	topDir := ""
	fi, err := os.Stat(asset)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		topDir = filepath.Dir(asset)
	}

	walkFn := func(inPath string, inInfo os.FileInfo, inErr error) error {
		if inErr != nil {
			return nil
		}

		path := filepath.Clean(inPath)
		baseName := filepath.Base(path)
		if topDir != "" {
			baseName = strings.TrimPrefix(path, topDir)
		}
		hostPath := filepath.Join(unitsDir, baseName)

		if inInfo.Mode().IsDir() {
			if inPath != asset {
				return filepath.SkipDir
			}
			_, err := os.Stat(hostPath)
			if err != nil && os.IsNotExist(err) {
				return os.MkdirAll(hostPath, 0755)
			}
			return err
		}

		if inInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			// This mimics `systemctl enable` behavior, expecting dependency
			// symlinks to be relative and pointing to units in the parent directory.
			linkDest, err := os.Readlink(path)
			if err != nil {
				return errors.Wrapf(err, "could not read link %q", path)
			}
			return os.Symlink(linkDest, hostPath)
		}

		if inInfo.Mode().IsRegular() {
			if _, err := os.Stat(hostPath); err == nil {
				// Do not overwrite previous assets
				return nil
			}
			return os.Symlink(path, hostPath)
		}
		return nil
	}

	return filepath.Walk(asset, walkFn)
}
