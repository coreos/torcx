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
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/torcx/internal/torcx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cmdProfileUse = &cobra.Command{
		Use:   "use-image --name|--path <IMNAME>:<REF>",
		Short: "adds an image to a profile",
		Long:  "adds the image referenced by NAME+REF to the list of images in profile PROFNAME",
		RunE:  runProfileUse,
	}
	flagProfileUseAllow string
	flagProfileUseName  string
	flagProfileUseFile  string
)

func init() {
	cmdProfile.AddCommand(cmdProfileUse)
	cmdProfileUse.Flags().StringVar(&flagProfileUseAllow, "allow", "", "pass --allow=missing to add a missing package to a profile")
	cmdProfileUse.Flags().StringVar(&flagProfileUseName, "name", "", "edit profile in user store with name NAME")
	cmdProfileUse.Flags().StringVar(&flagProfileUseFile, "file", "", "edit profile at path FILE")
}

func runProfileUse(cmd *cobra.Command, args []string) error {
	commonCfg, err := fillCommonRuntime("")
	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	if len(args) != 1 {
		return cmd.Usage()
	}

	imgPair := args[0]
	imstr := strings.SplitN(imgPair, ":", 2)
	if len(imstr) != 2 {
		return cmd.Usage()
	}

	imageName := imstr[0]
	imageRef := imstr[1]

	// Don't allow editing non-user profiles.
	profiles, err := torcx.ListProfiles([]string{commonCfg.UserProfileDir()})
	if err != nil {
		return errors.Wrap(err, "unable to list profiles")
	}

	if flagProfileUseName != "" {
		if flagProfileUseFile != "" {
			return cmd.Usage()
		}
		if _, ok := profiles[flagProfileUseName]; !ok {
			return fmt.Errorf("profile %s does not exist", flagProfileUseName)
		}

		flagProfileUseFile = filepath.Join(commonCfg.UserProfileDir(), flagProfileUseName+".json")
	}

	if flagProfileUseFile == "" {
		return cmd.Usage()
	}

	image := torcx.Image{
		Name:      imageName,
		Reference: imageRef,
	}

	//Scan the list of images, make sure it exists
	storeCache, err := torcx.NewStoreCache(commonCfg.StorePaths)
	if err != nil {
		return errors.Wrap(err, "unable to scan for packages")
	}
	if _, ok := storeCache.Images[image]; !ok {
		if flagProfileUseAllow == "missing" {
			logrus.WithFields(logrus.Fields{
				"image":     image.Name,
				"reference": image.Reference,
			}).Warn("Image does not exist, continuing")
		} else {
			return fmt.Errorf("Image %s does not exist, quitting. "+
				"(pass --allow=missing to force)", imgPair)
		}
	}

	if err := torcx.AddToProfile(flagProfileUseFile, image); err != nil {
		return errors.Wrap(err, "could not write new profile")
	}

	return nil
}
