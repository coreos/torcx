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
	cmdProfile = &cobra.Command{
		Use:   "profile [command]",
		Short: "Operate on local profile(s)",
		Long:  `This subcommand operates on local profile(s).`,
	}
)

func init() {
	TorcxCmd.AddCommand(cmdProfile)
}

// fillProfileRuntime generates the runtime config for profile subcommands,
// starting from system-wide state and config
func fillProfileRuntime(commonCfg *torcx.CommonConfig) (*torcx.ProfileConfig, error) {
	var (
		userProfileName   string
		vendorProfileName string
		curProfilePath    string
		nextProfile       string
	)

	if commonCfg == nil {
		return nil, errors.New("missing common configuration")
	}

	upn, vpn, err := torcx.CurrentProfileNames()
	if err == nil {
		userProfileName = upn
		vendorProfileName = vpn
	}
	cpp, err := torcx.CurrentProfilePath()
	if err == nil {
		curProfilePath = cpp
	}

	nextProfile, err = commonCfg.NextProfileName()
	if err != nil {
		nextProfile = ""
	}

	logrus.WithFields(logrus.Fields{
		"user profile":   userProfileName,
		"vendor profile": vendorProfileName,
		"next profile":   nextProfile,
	}).Debug("profile configuration parsed")

	return &torcx.ProfileConfig{
		CommonConfig:       *commonCfg,
		UserProfileName:    userProfileName,
		VendorProfileName:  vendorProfileName,
		CurrentProfilePath: curProfilePath,
		NextProfile:        nextProfile,
	}, nil
}
