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
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// ValidateCommonConfig performs validation on torcx common config
func ValidateCommonConfig(commonCfg *CommonConfig) error {
	if commonCfg == nil {
		return errors.New("nil common config")
	}
	if !filepath.IsAbs(commonCfg.BaseDir) {
		return errors.Errorf("non-absolute base_dir %q", commonCfg.BaseDir)
	}
	if !filepath.IsAbs(commonCfg.RunDir) {
		return errors.Errorf("non-absolute run_dir %q", commonCfg.RunDir)
	}
	if !filepath.IsAbs(commonCfg.ConfDir) {
		return errors.Errorf("non-absolute conf_dir %q", commonCfg.ConfDir)
	}
	for _, p := range commonCfg.StorePaths {
		if !filepath.IsAbs(p) {
			return errors.Errorf("non absolute store path %q", p)
		}
	}

	return nil
}

// ReadCommonConfig populates common config entries from optional settings from a config file
func ReadCommonConfig(cfgPath string, commonCfg *CommonConfig) error {
	if cfgPath == "" {
		// No config file, skip this
		return nil
	}
	if commonCfg == nil {
		return errors.New("nil common configuration")
	}

	fp, err := os.Open(cfgPath)
	if err != nil {
		// Missing config file, skip this
		return nil
	}
	defer fp.Close()
	fileCfg := ConfigV0{}
	if err := json.NewDecoder(bufio.NewReader(fp)).Decode(&fileCfg); err != nil {
		return err
	}
	if fileCfg.Kind != CommonConfigV0K {
		return errors.Errorf("unknown config file kind %q", fileCfg.Kind)
	}

	logrus.WithFields(logrus.Fields{
		"path": cfgPath,
	}).Debug("common config file read")

	// Populate with non-empty settings
	if fileCfg.Value.BaseDir != "" {
		commonCfg.BaseDir = fileCfg.Value.BaseDir
	}
	if fileCfg.Value.RunDir != "" {
		commonCfg.RunDir = fileCfg.Value.RunDir
	}
	if fileCfg.Value.ConfDir != "" {
		commonCfg.ConfDir = fileCfg.Value.ConfDir
	}
	if len(fileCfg.Value.StorePaths) > 0 {
		commonCfg.StorePaths = append(commonCfg.StorePaths, fileCfg.Value.StorePaths...)
	}

	return nil
}

// RuntimeConfigPath determines runtime location of torcx common configuration file.
func RuntimeConfigPath() string {
	cfgPath, err := procConfigPath()
	if err != nil {
		cfgPath = defaultCfgPath
	}
	return cfgPath
}

// procConfigPath parses `/proc/cmdline` looking for a `torcx_config=` boot-time option.
func procConfigPath() (string, error) {
	configKey := "torcx_config="

	bytes, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return "", err
	}
	for _, token := range strings.Fields(string(bytes)) {
		if strings.HasPrefix(token, configKey) {
			return strings.TrimPrefix(token, configKey), nil
		}
	}

	return "", errors.New("no custom torcx_config setting")
}
