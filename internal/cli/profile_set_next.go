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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/coreos/torcx/pkg/torcx"
)

var (
	cmdProfileSetNext = &cobra.Command{
		Use:   "set-next <NAME>",
		Short: "switches active profiles",
		Long:  "marks a given profile active for the next boot",
		RunE:  runProfileSetNext,
	}
)

func init() {
	cmdProfile.AddCommand(cmdProfileSetNext)
}

func runProfileSetNext(cmd *cobra.Command, args []string) error {
	commonCfg, err := fillCommonRuntime("")
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	if len(args) != 1 {
		return cmd.Usage()
	}

	profileName := args[0]

	profiles, err := torcx.ListProfiles(commonCfg.ProfileDirs())
	if err != nil {
		return errors.Wrap(err, "could not list profiles")
	}

	_, ok := profiles[profileName]
	if !ok {
		return fmt.Errorf("profile %s does not exist", profileName)
	}

	// This isn't quite atomic. It would be nice if it were
	err = commonCfg.SetNextProfileName(profileName)
	if err != nil {
		return errors.Wrap(err, "could not write profile file")
	}
	return nil
}
