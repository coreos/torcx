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
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/coreos/torcx/pkg/torcx"
)

var (
	cmdApply = &cobra.Command{
		Use:   "apply",
		Short: "apply a profile",
		RunE:  runApply,
	}
)

func init() {
	TorcxCmd.AddCommand(cmdApply)
}

func runApply(cmd *cobra.Command, args []string) error {
	var err error

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

// fillApplyRuntime generate runtime config for apply subcommand starting from
// system-wide configuration
func fillApplyRuntime(commonCfg *torcx.CommonConfig) (*torcx.ApplyConfig, error) {
	// If we fail to read lower profiles, report the error proceed without
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
