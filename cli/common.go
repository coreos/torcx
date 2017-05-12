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
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/coreos/torcx/pkg/torcx"
)

// fillCommonRuntime initializes common configuration settings, from several
// input sources (in order: defaults, file config, environment overrides).
func fillCommonRuntime() (*torcx.CommonConfig, error) {
	// Default common config settings
	commonCfg := torcx.CommonConfig{
		BaseDir: torcx.DefaultBaseDir,
		RunDir:  torcx.DefaultRunDir,
		ConfDir: torcx.DefaultConfDir,
		StorePaths: []string{
			torcx.VendorStorePath,
		},
	}

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

	// Add user and runtime store paths
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
