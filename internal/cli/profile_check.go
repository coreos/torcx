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

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/coreos/torcx/internal/torcx"
)

var (
	cmdProfileCheck = &cobra.Command{
		Use:   "check",
		Short: "check the manifest content and images for a profile",
		Long:  "Checks that the given profile (or the next profile on boot, if none is specified) is valid. Ensures that the schema is correct, and that all referenced images exist in the store",
		RunE:  runProfileCheck,
	}

	flagProfileCheckName      string
	flagProfileCheckPath      string
	flagProfileCheckOsVersion string
)

func init() {
	cmdProfile.AddCommand(cmdProfileCheck)
	cmdProfileCheck.Flags().StringVar(&flagProfileCheckName, "name", "", "profile name to check")
	cmdProfileCheck.Flags().StringVar(&flagProfileCheckPath, "file", "", "profile file to check")
	cmdProfileCheck.Flags().StringVarP(&flagProfileCheckOsVersion, "os-release", "n", "", "override OS version")
}

func runProfileCheck(cmd *cobra.Command, args []string) error {
	var err error

	commonCfg, err := fillCommonRuntime(flagProfileCheckOsVersion)
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	if len(args) != 0 {
		return cmd.Usage()
	}

	if flagProfileCheckPath == "" {
		if flagProfileCheckName == "" {
			flagProfileCheckName, err = commonCfg.NextProfileName()
			if err != nil {
				return errors.Wrapf(err, "unable to determine next profile")
			}

			logrus.Infof("No profile specified, using next profile %q", flagProfileCheckName)

			if flagProfileCheckName == torcx.VendorProfileName {
				logrus.Warn("Checking default (%s) profile - do you mean to do that?", flagProfileCheckName)
			}
		}

		localProfiles, err := torcx.ListProfiles(commonCfg.ProfileDirs())
		if err != nil {
			return errors.Wrap(err, "profiles listing failed")
		}

		var ok bool
		flagProfileCheckPath, ok = localProfiles[flagProfileCheckName]

		if !ok {
			return fmt.Errorf("profile %q not found", flagProfileCheckName)
		}
	}

	profile, err := torcx.ReadProfilePath(flagProfileCheckPath)
	if err != nil {
		return err
	}

	// Empty profiles are allowed
	if len(profile.Images) == 0 {
		logrus.Warn("Profile specifies no images")
		return nil
	}

	storeCache, err := torcx.NewStoreCache(commonCfg.StorePaths)
	if err != nil {
		return err
	}

	missing := false
	for _, im := range profile.Images {
		ar, err := storeCache.ArchiveFor(im)
		if err != nil {
			missing = true
			logrus.WithFields(logrus.Fields{
				"name":      im.Name,
				"reference": im.Reference,
			}).Error("image/reference not found")
		} else {
			logrus.WithFields(logrus.Fields{
				"name":         im.Name,
				"references":   im.Reference,
				"archive path": ar.Filepath,
			}).Debug("image/reference found")
		}
	}

	if missing {
		return fmt.Errorf("incomplete profile")
	}

	return nil
}
