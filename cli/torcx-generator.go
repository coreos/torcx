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
	"log/syslog"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/coreos/torcx/pkg/torcx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	// TorcxGenCmd is the top-level cobra command for `torcx-generator`
	TorcxGenCmd = &cobra.Command{
		Use:          "torcx-generator",
		RunE:         runGenerator,
		SilenceUsage: true,
	}
)

func runGenerator(cmd *cobra.Command, args []string) error {
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err == nil {
		logrus.AddHook(hook)
	}

	// TODO(lucab): provide config via persistent JSON file and /proc/cmdline
	logrus.SetLevel(logrus.DebugLevel)

	commonCfg, err := fillCommonRuntime()
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}
	if torcx.IsExistingPath(commonCfg.RunDir) {
		logrus.Info("torcx already run")
		return nil
	}

	applyCfg, err := fillApplyRuntime(commonCfg)
	if err != nil {
		return errors.Wrap(err, "apply configuration failed")
	}

	err = torcx.ApplyProfile(applyCfg)
	if err != nil {
		return errors.Wrap(err, "apply failed")
	}

	err = torcx.SealSystemState(applyCfg)
	if err != nil {
		return errors.Wrapf(err, "sealing system state failed")
	}
	return nil
}

// fillApplyRuntime populates runtime config starting from system-wide configuration
func fillApplyRuntime(commonCfg *torcx.CommonConfig) (*torcx.ApplyConfig, error) {
	lowerProfileNames, err := lowerProfiles(commonCfg)
	if err != nil {
		return nil, err
	}

	// If we fail to read /etc/torcx/next-profile, report the error and proceed without
	upperProfileName, err := commonCfg.NextProfileName()
	if err != nil {
		logrus.Warnf("no next profile: %s", err)
		upperProfileName = ""
	}

	logrus.WithFields(logrus.Fields{
		"lower profiles (vendor/oem)": lowerProfileNames,
		"upper profile (user)":        upperProfileName,
	}).Debug("apply configuration parsed")

	return &torcx.ApplyConfig{
		CommonConfig:  *commonCfg,
		LowerProfiles: lowerProfileNames,
		UpperProfile:  upperProfileName,
	}, nil
}

// lowerProfiles returns a list of lower profiles (vendor/oem) found on
// the system at runtime. The set may be empty.
func lowerProfiles(commonCfg *torcx.CommonConfig) ([]string, error) {
	if commonCfg == nil {
		return nil, errors.New("missing common configuration")
	}

	lowerProfiles := []string{}
	localProfiles, err := torcx.ListProfiles(commonCfg.ProfileDirs())
	if err != nil {
		return nil, err
	}
	for _, prof := range torcx.DefaultLowerProfiles {
		if path, ok := localProfiles[prof]; ok && path != "" {
			lowerProfiles = append(lowerProfiles, prof)
		} else {
			logrus.WithFields(logrus.Fields{
				"missing profile": prof,
			}).Debug("skipped missing lower profile")
		}
	}
	return lowerProfiles, nil
}
