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
	"reflect"
	"testing"
)

func TestFilterStoreVersion(t *testing.T) {
	tests := []struct {
		desc          string
		stores        []string
		curVersion    string
		filterVersion string

		resultStore []string
	}{
		{
			"empty",
			[]string{},
			"foo",
			"bar",
			[]string{},
		},
		{
			"matching version",
			[]string{VendorStoreDir(VendorUsrDir)},
			"1.0.0",
			"1.0.0",
			[]string{VendorStoreDir(VendorUsrDir)},
		},
		{
			"non-matching version",
			[]string{VendorStoreDir(VendorUsrDir)},
			"1.0.0",
			"2.0.0",
			[]string{},
		},
		{
			"unrelated, OEM",
			[]string{OemStoreDir},
			"1.0.0",
			"2.0.0",
			[]string{OemStoreDir},
		},
	}

	for _, tt := range tests {
		t.Logf("Checking %q", tt.desc)

		res := FilterStoreVersions(VendorUsrDir, tt.stores, tt.curVersion, tt.filterVersion)
		if !reflect.DeepEqual(res, tt.resultStore) {
			t.Fatalf("expected %#v, got %#v", tt.resultStore, res)
		}

	}
}
