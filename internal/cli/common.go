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

package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/coreos/torcx/internal/torcx"
)

// fillCommonRuntime initializes common configuration settings, from several
// input sources (in order: defaults, file config, environment overrides).
func fillCommonRuntime(OsRelease string) (*torcx.CommonConfig, error) {
	var err error

	usrMountpoint := torcx.VendorUsrDir
	path, ok := viper.Get("USR_MOUNTPOINT").(string)
	if ok && filepath.IsAbs(path) {
		usrMountpoint = path
	}

	// Default common config settings
	commonCfg := torcx.CommonConfig{
		BaseDir: torcx.DefaultBaseDir,
		RunDir:  torcx.DefaultRunDir,
		UsrDir:  usrMountpoint,
		ConfDir: torcx.DefaultConfDir,
		StorePaths: []string{
			torcx.VendorStoreDir(usrMountpoint),
		},
	}

	// Determine OS release ID
	if OsRelease == "" {
		path := torcx.VendorOsReleasePath(usrMountpoint)
		OsRelease, err = torcx.CurrentOsVersionID(path)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"path": path,
			}).Warn("unable to detect OS version-id")
		}
	}

	// Add OEM store (versioned first)
	if OsRelease != "" {
		commonCfg.StorePaths = append(commonCfg.StorePaths, filepath.Join(torcx.OemStoreDir, OsRelease))
	}
	commonCfg.StorePaths = append(commonCfg.StorePaths, torcx.OemStoreDir)

	// Read common config from config file, if present
	cfgPath := torcx.RuntimeConfigPath()
	if err := torcx.ReadCommonConfig(cfgPath, &commonCfg); err != nil {
		return nil, errors.Wrapf(err, "reading common config from %q", cfgPath)
	}

	// Overrides from environment
	if baseDir := viper.GetString("basedir"); baseDir != "" {
		commonCfg.BaseDir = baseDir
	}
	if rundir := viper.GetString("rundir"); rundir != "" {
		commonCfg.RunDir = rundir
	}
	if confdir := viper.GetString("confdir"); confdir != "" {
		commonCfg.ConfDir = confdir
	}

	// Add user and runtime store paths (versioned first)
	if OsRelease != "" {
		commonCfg.StorePaths = append(commonCfg.StorePaths, filepath.Join(commonCfg.BaseDir, "store", OsRelease))
	}
	commonCfg.StorePaths = append(commonCfg.StorePaths, filepath.Join(commonCfg.BaseDir, "store"))
	extraStorePaths := viper.GetStringSlice("storepath")
	if extraStorePaths != nil {
		commonCfg.StorePaths = append(commonCfg.StorePaths, extraStorePaths...)
	}

	if err := torcx.ValidateCommonConfig(&commonCfg); err != nil {
		return nil, errors.Wrap(err, "invalid common config")
	}
	logrus.WithFields(logrus.Fields{
		"base_dir":    commonCfg.BaseDir,
		"run_dir":     commonCfg.RunDir,
		"conf_dir":    commonCfg.ConfDir,
		"store_paths": commonCfg.StorePaths,
	}).Debug("common configuration parsed")

	return &commonCfg, nil
}

// hasExpFeature checks if an experimental feature is enabled
// via its corresponding `TORCX_EXP_<featureName>` env flag.
func hasExpFeature(featureName string) bool {
	featureBase := "TORCX_EXP_"
	feat := featureBase + strings.ToUpper(featureName)
	return os.Getenv(feat) != ""
}
