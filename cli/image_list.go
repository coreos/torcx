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

	"github.com/coreos/torcx/pkg/torcx"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	cmdImageList = &cobra.Command{
		Use:   "list [INAME]",
		Short: "list available images and references",
		Long: `List all images in the stores, as well as available references.
If "INAME" is specified, only list the references for that image name.`,
		RunE: runImageList,
	}
	flagImageListOsVersion string
)

func init() {
	cmdImage.AddCommand(cmdImageList)
	cmdImageList.Flags().StringVarP(&flagImageListOsVersion, "os-release", "n", "", "override OS version")
}

func runImageList(cmd *cobra.Command, args []string) error {
	var err error

	imageName := ""
	if len(args) > 1 {
		return errors.New("too many arguments")
	}
	if len(args) == 1 {
		imageName = args[0]
	}

	commonCfg, err := fillCommonRuntime(flagImageListOsVersion)

	if err != nil {
		return errors.Wrap(err, "common configuration failed")
	}

	storeCache, err := torcx.NewStoreCache(commonCfg.StorePaths)
	if err != nil {
		return err
	}

	imgList := make([]imageEntry, 0, len(storeCache.Images))
	for _, arch := range storeCache.Images {
		if imageName != "" && arch.Name != imageName {
			continue
		}
		imgList = append(imgList, imageEntry{arch})

		// TODO(caseyc): inspect xattr metadata
	}

	imageListOut := ImageList{
		Kind:  TorcxImageListV0K,
		Value: imgList,
	}

	jsonOut := json.NewEncoder(os.Stdout)
	jsonOut.SetIndent("", "  ")
	err = jsonOut.Encode(imageListOut)

	return err
}
