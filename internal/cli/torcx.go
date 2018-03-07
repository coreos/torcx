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
	"github.com/coreos/torcx/pkg/multicall"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// GlobalCfg is for global CLI state
type GlobalCfg struct {
	// logrus has its own global state for verbosity level
}

var (
	// TorcxCmd is the top-level cobra command for `torcx`
	TorcxCmd = &cobra.Command{
		Use:           "torcx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// TorcxCliCfg holds global CLI status available to all subcommands
	TorcxCliCfg GlobalCfg
)

// Init initializes the CLI environment for torcx
func Init() error {
	viper.SetEnvPrefix("TORCX")
	viper.AutomaticEnv()

	logrus.SetLevel(logrus.WarnLevel)

	verboseFlag := TorcxCmd.PersistentFlags().VarPF((*cliCfgVerbose)(&TorcxCliCfg), "verbose", "v", "verbosity level")
	verboseFlag.NoOptDefVal = "info"

	multicall.AddCobra(TorcxCmd.Use, TorcxCmd)
	multicall.AddCobra(TorcxGenCmd.Use, TorcxGenCmd)

	return nil
}

// cliCfgVerbose is for `--verbose` flag
type cliCfgVerbose GlobalCfg

func (ccv *cliCfgVerbose) Set(s string) error {
	lvl, err := logrus.ParseLevel(s)
	if err != nil {
		return err
	}

	logrus.SetLevel(lvl)
	return nil
}

func (ccv *cliCfgVerbose) String() string {
	return logrus.GetLevel().String()
}

func (ccv *cliCfgVerbose) Type() string {
	return "level"
}
