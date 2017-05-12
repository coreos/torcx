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
	"log/syslog"

	"github.com/Sirupsen/logrus"
	logrus_syslog "github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/spf13/cobra"
)

var (
	// TorcxGenCmd is the top-level cobra command for `torcx-generator`
	TorcxGenCmd = &cobra.Command{
		Use:          "torcx-generator",
		RunE:         runGenerator,
		SilenceUsage: true,
	}
)

func runGenerator(cmd *cobra.Command, args []string) error {
	hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, "")
	if err == nil {
		logrus.AddHook(hook)
	}

	// TODO(lucab): provide config via persistent JSON file and /proc/cmdline
	logrus.SetLevel(logrus.DebugLevel)

	return runApply(cmd, args)
}
