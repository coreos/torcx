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

func fillCommonRuntime() (*torcx.CommonConfig, error) {
	baseDir := viper.GetString("basedir")
	if baseDir == "" {
		baseDir = "/var/lib/torcx"
	}
	if !filepath.IsAbs(baseDir) {
		return nil, errors.New("non-absolute basedir")
	}

	rundir := viper.GetString("rundir")
	if rundir == "" {
		rundir = "/var/run/torcx"
	}
	if !filepath.IsAbs(rundir) {
		return nil, errors.New("non-absolute rundir")
	}

	confdir := viper.GetString("confdir")
	if confdir == "" {
		confdir = "/etc/torcx"
	}
	if !filepath.IsAbs(confdir) {
		return nil, errors.New("non-absolute confdir")
	}

	storePaths := []string{
		filepath.Join(torcx.VENDOR_DIR, "store"),
		filepath.Join(baseDir, "store"), // the user store path
	}
	extraStorePaths := viper.GetStringSlice("storepath")
	if extraStorePaths != nil {
		storePaths = append(storePaths, extraStorePaths...)
	}

	logrus.WithFields(logrus.Fields{
		"basedir":     baseDir,
		"rundir":      rundir,
		"confdir":     confdir,
		"store paths": storePaths,
	}).Debug("common configuration parsed")

	return &torcx.CommonConfig{
		BaseDir:    baseDir,
		RunDir:     rundir,
		ConfDir:    confdir,
		StorePaths: storePaths,
	}, nil
}
