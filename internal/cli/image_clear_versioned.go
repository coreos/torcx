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

package cli

import (
	"os"
	"path/filepath"

	"github.com/coreos/torcx/internal/torcx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cmdImageClearVersioned = &cobra.Command{
		Use:   "clear-versioned [INAME]",
		Short: "clear images from versioned stores",
		Long: `Clear images from versioned stores.
If "INAME" is specified, only clear that image name from stores.`,
		RunE: runImageClearVersioned,
	}
	flagImageClearVersionedKeepVersions []string

	errNoStoreVersions = errors.New("no store versions to keep")
	errNilCommonConfig = errors.New("nil CommongConfig")
	errEmptyImageName  = errors.New("empty imageName")
)

func init() {
	cmdImage.AddCommand(cmdImageClearVersioned)
	cmdImageClearVersioned.Flags().StringArrayVarP(&flagImageClearVersionedKeepVersions, "keep-version", "k", []string{}, "store version to keep")
}

func runImageClearVersioned(cmd *cobra.Command, args []string) error {
	keepVersions := flagImageClearVersionedKeepVersions
	if len(keepVersions) == 0 {
		return errNoStoreVersions
	}
	if len(args) > 1 {
		return errors.New("too many arguments")
	}
	imageName := ""
	if len(args) == 1 {
		imageName = args[0]
	}

	commonCfg, err := fillCommonRuntime("")
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	if imageName != "" {
		if err := clearVersionedImage(commonCfg, imageName, keepVersions); err != nil {
			return errors.Wrapf(err, "failed to clear versioned image %s", imageName)
		}
	} else {
		if err := clearVersionedStores(commonCfg, keepVersions); err != nil {
			return errors.Wrapf(err, "failed to clear versioned stores")
		}
	}

	return nil
}

// clearVersionedImage deletes all references for the image `imageName` from all
// versioned stores, except those listed in `keepVersions`.
func clearVersionedImage(cfg *torcx.CommonConfig, imageName string, keepVersions []string) error {
	if cfg == nil {
		return errNilCommonConfig
	}
	if imageName == "" {
		return errEmptyImageName
	}

	globPattern := filepath.Join(cfg.UserStorePath("*"), imageName+":*.torcx.*")
	images, err := filepath.Glob(globPattern)
	if err != nil {
		return errors.Wrapf(err, "failed to glob on %s", globPattern)
	}

	removals := []string{}
	for _, im := range images {
		imPath := filepath.Clean(im)
		fi, err := os.Stat(imPath)
		if err != nil {
			continue
		}
		if !fi.Mode().IsRegular() {
			continue
		}

		storePath := filepath.Dir(imPath)
		if !shouldKeep(cfg, storePath, keepVersions) {
			removals = append(removals, imPath)
		}
	}

	for _, imPath := range removals {
		logrus.WithFields(logrus.Fields{
			"path": imPath,
		}).Info("removing image from versioned store")
		if err := os.Remove(imPath); err != nil {
			return err
		}
	}

	return nil
}

// clearVersionedStores deletes all versioned stores, except those listed in `keepVersions`.
func clearVersionedStores(cfg *torcx.CommonConfig, keepVersions []string) error {
	if cfg == nil {
		return errNilCommonConfig
	}

	globPattern := cfg.UserStorePath("*")
	stores, err := filepath.Glob(globPattern)
	if err != nil {
		return errors.Wrapf(err, "failed to glob on %s", globPattern)
	}

	removals := []string{}
	for _, store := range stores {
		storePath := filepath.Clean(store)
		fi, err := os.Stat(storePath)
		if err != nil {
			continue
		}
		if !fi.Mode().IsDir() {
			continue
		}

		if !shouldKeep(cfg, storePath, keepVersions) {
			removals = append(removals, storePath)
		}
	}

	for _, storePath := range removals {
		logrus.WithFields(logrus.Fields{
			"path": storePath,
		}).Info("removing versioned store")
		if err := os.RemoveAll(storePath); err != nil {
			return err
		}
	}

	return nil
}

// shouldKeep returns whether `storePath` is a versioned store to keep,
// according to `keepVersions`.
func shouldKeep(cfg *torcx.CommonConfig, storePath string, keepVersions []string) bool {
	for _, ver := range keepVersions {
		verPath := filepath.Clean(cfg.UserStorePath(ver))
		if storePath == "" ||
			verPath == "" ||
			storePath == verPath {
			return true
		}
	}
	return false
}
