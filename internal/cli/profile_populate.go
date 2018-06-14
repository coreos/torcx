// Copyright 2018 CoreOS Inc.
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
	"context"
	"fmt"
	"os"
	"time"

	"github.com/coreos/torcx/internal/torcx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	cmdProfilePopulate = &cobra.Command{
		Use:   "populate",
		Short: "Populates the store with all image required by the given profile",
		RunE:  runProfilePopulate,
	}

	flagProfilePopulateName      string
	flagProfilePopulatePath      string
	flagProfilePopulateOsVersion string

	// TODO(lucab): consider whether to make this configurable
	timeoutMins = 1
)

func init() {
	cmdProfile.AddCommand(cmdProfilePopulate)
	cmdProfilePopulate.Flags().StringVar(&flagProfilePopulateName, "name", "", "profile name to populate")
	cmdProfilePopulate.Flags().StringVar(&flagProfilePopulatePath, "file", "", "profile file to populate")
	cmdProfilePopulate.Flags().StringVarP(&flagProfilePopulateOsVersion, "os-release", "n", "", "override OS version")
}

func runProfilePopulate(cmd *cobra.Command, args []string) error {
	commonCfg, err := fillCommonRuntime(flagProfilePopulateOsVersion)
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	if len(args) != 0 {
		return cmd.Usage()
	}

	if flagProfilePopulatePath == "" {
		if flagProfilePopulateName == "" {
			flagProfilePopulateName, err = commonCfg.NextProfileName()
			if err != nil {
				return errors.Wrapf(err, "unable to determine next profile")
			}

			logrus.Infof("using next profile %q", flagProfilePopulateName)

			if flagProfilePopulateName == torcx.VendorProfileName {
				logrus.Warn("using default profile (%s), which should not require external images", torcx.VendorProfileName)
			}
		}

		localProfiles, err := torcx.ListProfiles(commonCfg.ProfileDirs())
		if err != nil {
			return errors.Wrap(err, "profiles listing failed")
		}

		var ok bool
		flagProfilePopulatePath, ok = localProfiles[flagProfilePopulateName]

		if !ok {
			return fmt.Errorf("profile %q not found", flagProfilePopulateName)
		}
	}

	profile, err := torcx.ReadProfilePath(flagProfilePopulatePath)
	if err != nil {
		return err
	}

	// Empty profiles are allowed
	if len(profile) == 0 {
		logrus.Warn("profile specifies no images")
		return nil
	}

	storeCache, err := torcx.NewStoreCache(commonCfg.StorePaths)
	if err != nil {
		return err
	}

	remotes := []string{}
	{
		keys := map[string]bool{}
		for _, im := range profile {
			if im.Remote == "" {
				continue
			}
			if ok := keys[im.Remote]; ok {
				continue
			}
			remotes = append(remotes, im.Remote)
			keys[im.Remote] = true
		}
	}

	// Nothing to fetch
	if len(remotes) <= 0 {
		logrus.Warn("profile references no remote images")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMins)*time.Minute)
	defer cancel()
	remotesCache, err := torcx.NewRemotesCache(ctx, commonCfg.UsrDir, commonCfg.RemotesDirs(), remotes)
	if err != nil {
		return err
	}

	versionedStorePath := commonCfg.UserStorePath(flagProfilePopulateOsVersion)
	if err := os.MkdirAll(versionedStorePath, 0755); err != nil {
		return err
	}

	localCount := 0
	remoteCount := 0
	// TODO(lucab): parallelize this
	for _, im := range profile {
		if archive, err := storeCache.ArchiveFor(im); err == nil {
			logrus.WithFields(logrus.Fields{
				"path": archive.Filepath,
			}).Info("image found locally")
			localCount++
			continue
		}

		ctxTo, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMins)*time.Minute)
		defer cancel()
		if err := remotesCache.FetchImage(ctxTo, im, versionedStorePath); err != nil {
			return err
		}
		remoteCount++
	}

	logrus.WithFields(logrus.Fields{
		"local":        localCount,
		"downloaded":   remoteCount,
		"profile_name": flagProfilePopulateName,
		"profile_path": flagProfilePopulatePath,
	}).Info("store populated")
	return nil
}
