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
	"strings"
	"testing"
)

func TestGetOsReleaseID(t *testing.T) {
	tests := []struct {
		desc    string
		content string

		expVer string
		expErr error
	}{
		{
			"empty",
			"",

			"",
			ErrUnknownOSVersionID,
		},
		{
			"positive",
			"VERSION_ID=1.0.0",

			"1.0.0",
			nil,
		},
		{
			"positive with empty lines",
			"\nVERSION_ID=1.0.0\n",

			"1.0.0",
			nil,
		},
		{
			"missing key",
			"=1.0.0",

			"",
			ErrUnknownOSVersionID,
		},
		{
			"missing value",
			"VERSION_ID=",

			"",
			ErrUnknownOSVersionID,
		},
	}

	for _, tt := range tests {
		rd := strings.NewReader(tt.content)
		ver, err := parseOsVersionID(rd)
		if err != tt.expErr {
			t.Errorf("testcase %q failed with mismatched error:\n got: %v\n expected: %v", tt.desc, err, tt.expErr)
		}
		if tt.expErr != nil {
			continue
		}
		if ver != tt.expVer {
			t.Fatalf("testcase %q failed with mismatched version-id result: got %q - expected %q", tt.desc, err, tt.expVer)
		}
	}

}
