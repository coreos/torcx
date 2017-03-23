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
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/coreos/torcx/pkg/torcx"
)

var (
	cmdProfileCheck = &cobra.Command{
		Use:   "check <PNAME>",
		Short: "check the manifest content and bundles for a profile",
		Long:  "check the manifest content and bundles for profile named <PNAME>",
		RunE:  runProfileCheck,
	}
)

func init() {
	cmdProfile.AddCommand(cmdProfileCheck)
}

func runProfileCheck(cmd *cobra.Command, args []string) error {
	var err error

	if len(args) != 1 {
		return errors.New("missing profile name")
	}
	profileName := args[0]

	commonCfg, err := fillCommonRuntime()
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}
	profileCfg, err := fillProfileRuntime(commonCfg)
	if err != nil {
		return errors.Wrap(err, "profile configuration failed")
	}

	profileDirs := []string{
		filepath.Join(torcx.VENDOR_DIR, "profiles.d"),
		filepath.Join(profileCfg.ConfDir, "profiles.d"),
	}
	localProfiles, err := torcx.ListProfiles(profileDirs)
	if err != nil {
		return errors.Wrap(err, "profiles listing failed")
	}

	path, ok := localProfiles[profileName]
	if !ok {
		return fmt.Errorf("profile %q not found", args[0])
	}

	bundles, err := torcx.ReadProfile(path)
	if err != nil {
		return err
	}
	if len(bundles.Archives) == 0 {
		return nil
	}

	storeCache, err := torcx.NewStoreCache(profileCfg.StorePaths)
	if err != nil {
		return err
	}

	incomplete := false
	for _, pkg := range bundles.Archives {
		_, err := storeCache.LookupReference(pkg)
		if err != nil {
			incomplete = true
			logrus.WithFields(logrus.Fields{
				"bundle name": pkg.Image,
				"reference":   pkg.Reference,
			}).Error("bundle/reference not found")
		}
	}
	if incomplete {
		return fmt.Errorf("incomplete profile")
	}

	return nil
}
