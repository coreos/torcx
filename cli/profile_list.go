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
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/coreos/torcx/pkg/torcx"
)

var (
	cmdProfileList = &cobra.Command{
		Use:   "list",
		Short: "list available profiles",
		RunE:  runProfileList,
	}
)

func init() {
	cmdProfile.AddCommand(cmdProfileList)
}

func runProfileList(cmd *cobra.Command, args []string) error {
	var err error

	commonCfg, err := fillCommonRuntime()
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}
	profileCfg, err := fillProfileRuntime(commonCfg)
	if err != nil {
		return errors.Wrap(err, "profile configuration failed")
	}

	localProfiles, err := torcx.ListProfiles(commonCfg.ProfileDirs())
	if err != nil {
		return errors.Wrap(err, "profiles listing failed")
	}
	profNames := make([]string, 0, len(localProfiles))
	for k := range localProfiles {
		profNames = append(profNames, k)
	}

	var userName, vendorName, nextName, curPath *string
	if profileCfg.UserProfileName != "" {
		userName = &profileCfg.UserProfileName
	}
	if profileCfg.VendorProfileName != "" {
		vendorName = &profileCfg.VendorProfileName
	}
	if profileCfg.NextProfile != "" {
		nextName = &profileCfg.NextProfile
	}
	if profileCfg.CurrentProfilePath != "" {
		curPath = &profileCfg.CurrentProfilePath
	}

	profListOut := ProfileList{
		Kind: TorcxProfileListV0K,
		Value: profileList{
			UserProfileName:    userName,
			VendorProfileName:  vendorName,
			CurrentProfilePath: curPath,
			NextProfileName:    nextName,
			Profiles:           profNames,
		},
	}

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")
	err = jsonOut.Encode(profListOut)

	return err
}
