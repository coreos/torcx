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
	"strings"

	"github.com/pkg/errors"
)

// ReadMetadata returns metadata regarding the currently running profile,
// as read from the metadata file
func ReadMetadata(fusePath string) (map[string]string, error) {
	meta := make(map[string]string)

	if !IsSystemSealed(fusePath) {
		return meta, errors.New("no active profile")
	}

	fp, err := os.Open(fusePath)
	if err != nil {
		return meta, err
	}
	defer fp.Close()

	sc := bufio.NewScanner(fp)
	for sc.Scan() {
		line := sc.Text()
		tokens := strings.SplitN(line, "=", 2)
		if len(tokens) == 2 {
			meta[tokens[0]] = strings.Trim(tokens[1], `"`)
		}
	}

	return meta, nil
}

// IsSystemSealed checks whether the runtime seal is already applied
func IsSystemSealed(fusePath string) bool {
	_, err := os.Lstat(fusePath)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
