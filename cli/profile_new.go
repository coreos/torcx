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
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/coreos/torcx/pkg/torcx"
)

var (
	cmdProfileNew = &cobra.Command{
		Use:   "new <NAME>",
		Short: "creates a new profile",
		Long:  "Creates a new profile, optionally from an existing profile",
		RunE:  runProfileNew,
	}
	flagProfileNewFromNext bool
	flagProfileNewFrom     string
	flagProfileNewName     string
	flagProfileNewFile     string
)

func init() {
	cmdProfile.AddCommand(cmdProfileNew)
	cmdProfileNew.Flags().StringVar(&flagProfileNewFrom, "from", "", "profile name to duplicate")
	cmdProfileNew.Flags().BoolVar(&flagProfileNewFromNext, "from-next", false, "duplicate profile marked for boot")
	cmdProfileNew.Flags().StringVar(&flagProfileNewName, "name", "", "create profile in the store with name NAME")
	cmdProfileNew.Flags().StringVar(&flagProfileNewFile, "file", "", "create profile at path FILE")

}

func runProfileNew(cmd *cobra.Command, args []string) error {
	commonCfg, err := fillCommonRuntime()
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	profiles, err := torcx.ListProfiles(commonCfg.ProfileDirs())
	if err != nil {
		return errors.Wrap(err, "Could not list profiles")
	}

	if flagProfileNewName != "" {
		if flagProfileNewFile != "" {
			return cmd.Usage()
		}

		if _, ok := profiles[flagProfileNewName]; ok {
			return fmt.Errorf("profile %s already exists", flagProfileNewName)
		}

		flagProfileNewFile = filepath.Join(commonCfg.UserProfileDir(), flagProfileNewName)
	}

	if flagProfileNewFile == "" {
		return cmd.Usage()
	}

	if flagProfileNewFromNext && flagProfileNewFrom != "" {
		return fmt.Errorf("error, --from and --from-next cannot both be specified")
	}

	if flagProfileNewFromNext {
		flagProfileNewFrom, err = commonCfg.NextProfileName()
		if err != nil {
			return errors.Wrap(err, "could not read next profile")
		}
	}

	if flagProfileNewFrom != "" {
		return copyProfile(profiles, flagProfileNewFrom, flagProfileNewFile)
	}

	return newBlankProfile(flagProfileNewFile)
}

func copyProfile(profiles map[string]string, fromName string, toPath string) error {
	fromPath, ok := profiles[fromName]
	if !ok {
		return fmt.Errorf("Could not find profile %s", fromName)
	}

	fromFp, err := os.Open(fromPath)
	if err != nil {
		return errors.Wrap(err, "could not open source profile")
	}
	defer fromFp.Close()

	toFp, err := os.Create(toPath)
	if err != nil {
		return errors.Wrap(err, "could not create destination profile")
	}
	defer toFp.Close()

	if _, err = io.Copy(toFp, fromFp); err != nil {
		return errors.Wrap(err, "could not write destination profile")
	}

	if err := toFp.Close(); err != nil {
		return errors.Wrap(err, "could not close destination profile")
	}

	return nil
}

func newBlankProfile(toPath string) error {
	blank := torcx.ProfileManifestV0{
		Kind: torcx.ProfileManifestV0K,
		Value: torcx.Images{
			Images: []torcx.Image{},
		},
	}

	fp, err := os.Create(toPath)
	if err != nil {
		return errors.Wrap(err, "could not create profile")
	}

	enc := json.NewEncoder(fp)

	if err := enc.Encode(blank); err != nil {
		return errors.Wrap(err, "could not write profile")
	}
	return nil
}
