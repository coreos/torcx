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

import "github.com/spf13/cobra"

var (
	cmdImage = &cobra.Command{
		Use:   "image [command]",
		Short: "Operate on images(s) in the local store",
		Long:  `This subcommand operates on image(s) in the local store.`,
	}
)

func init() {
	TorcxCmd.AddCommand(cmdImage)
}
