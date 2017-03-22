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
	"fmt"

	"encoding/json"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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
		return errors.Wrap(err, "apply configuration failed")
	}

	var cp *string
	if profileCfg.CurrentProfile != "" {
		cp = &profileCfg.CurrentProfile
	}

	profListOut := ProfileList{
		Kind: TorcxCliV0,
		Value: profileList{
			CurrentProfileName: cp,
			NextProfileName:    profileCfg.NextProfile,
		},
	}

	out, _ := json.MarshalIndent(profListOut, "", "    ")
	fmt.Print(string(out))

	return nil
}
