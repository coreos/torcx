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

package torcx

import (
	"bufio"
	"os"

	"github.com/pkg/errors"
	"strings"
)

type archive struct {
	Image string `json:"image"`
	Ref   string `json:"ref"`
}

// RunningProfile return the currently running profile
func RunningProfile() (string, error) {
	var profile string

	if !IsFuseBlown() {
		return "", errors.New("no active profile")
	}

	fp, err := os.Open(FUSE_PATH)
	if err != nil {
		return "", err
	}

	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		line := sc.Text()
		tokens := strings.SplitN(line, "", 2)
		if len(tokens) == 2 && tokens[0] == FUSE_PROFILE {
			profile = tokens[1]
			break
		}
	}

	if profile == "" {
		return "", errors.New("unable to read profile")
	}

	return profile, nil
}

func ReadProfile() ([]archive, error) {
	pkglist := []archive{}

	return pkglist, nil
}
