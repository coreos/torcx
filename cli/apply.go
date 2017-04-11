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
	applyCfg, err := fillApplyRuntime(commonCfg)
	if err != nil {
		return errors.Wrap(err, "apply configuration failed")
	}

	if torcx.IsSystemSealed(torcx.FUSE_PATH) {
		return errors.New("system already sealed")
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
	// We ignore the error here; NextProfileName will always return something we
	// can apply, but also returns errors for other uses
	profileName, err := commonCfg.NextProfileName()
	if err != nil {
		logrus.Warn("Falling back to default profile:", err)
		profileName = torcx.DEFAULT_PROFILE_NAME
	}

	logrus.WithFields(logrus.Fields{
		"profile": profileName,
	}).Debug("apply configuration parsed")

	return &torcx.ApplyConfig{
		CommonConfig: *commonCfg,
		Profile:      profileName,
	}, nil
}
