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
	var userProfileName, vendorProfileName string

	vendorProfileName = torcx.VendorProfileName
	// If we fail to read /etc/torcx/next-profile, report the error and use the default
	userProfileName, err := commonCfg.NextProfileName()
	if err != nil {
		logrus.Warnf("no next profile: %s", err)
		userProfileName = ""
	}

	logrus.WithFields(logrus.Fields{
		"vendor profile (lower)": vendorProfileName,
		"user profile (upper)":   userProfileName,
	}).Debug("apply configuration parsed")

	return &torcx.ApplyConfig{
		CommonConfig: *commonCfg,
		LowerProfile: vendorProfileName,
		UpperProfile: userProfileName,
	}, nil
}
