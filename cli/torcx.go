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
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// TorcxCmd is the top-level cobra command for torcx
	TorcxCmd = &cobra.Command{
		Use: "torcx",
	}
)

// Init initializes the CLI environment for torcx
func Init() error {
	viper.SetEnvPrefix("TORCX")
	viper.AutomaticEnv()
	logrus.SetLevel(logrus.DebugLevel)
	return nil
}
