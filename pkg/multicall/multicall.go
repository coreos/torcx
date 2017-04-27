// Copyright 2015 The rkt Authors
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

//+build linux

// Package multicall provides facilities to build a binary which
// behaves in different ways depending on how it has been invoked.
// It integrates directly with cobra, supporting CLI binaries
// with sub-commands.
package multicall

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
)

var commands = make(map[string]*cobra.Command)

// AddCobra adds a new multicall name backed by a cobra command.
// `name` is the command name and `cmd` is the cobra command that
// will be executed for the specified named command.
func AddCobra(name string, cmd *cobra.Command) error {
	if name == "" {
		return fmt.Errorf("empty multicall name provided")
	}

	if cmd == nil {
		return fmt.Errorf("invalid multicall function provided")
	}

	if _, ok := commands[name]; ok {
		return fmt.Errorf("command with name %q already exists", name)
	}

	commands[name] = cmd
	return nil
}

// switchMulticall returns the cobra command corresponding to multicall binary `name`.
func switchMulticall(name string) (*cobra.Command, error) {
	if name == "" {
		return nil, fmt.Errorf("empty multicall name provided")
	}

	cmd, ok := commands[name]
	if !ok {
		return nil, fmt.Errorf("unknown multicall name %q", name)
	}

	return cmd, nil
}

// getName returns base name of the process executable, optionally
// trying to resolve links.
func getName(followLinks bool) string {
	name := os.Args[0]

	if followLinks {
		exePath, err := os.Readlink("/proc/self/exe")
		if err == nil {
			name = exePath
		}
	}

	return path.Base(name)
}

// MultiExecute dispatches execution based on binary name, optionally
// resolving links.
func MultiExecute(followLinks bool) error {
	cliName := getName(followLinks)
	cmd, err := switchMulticall(cliName)
	if err != nil {
		return err
	}

	return cmd.Execute()
}
