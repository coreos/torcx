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
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var inManifest = ProfileManifestV0{
	Kind: "profile-manifest-v0",
	Value: Images{
		[]Image{
			Image{
				Name:      "test-name",
				Reference: "test-reference",
			},
		},
	},
}

func TestGetProfile(t *testing.T) {
	// Schema of profile v0 is described in
	// https://github.com/coreos/torcx/blob/master/Documentation/schemas/profile-manifest-v0.md
	profilePath := "../../fixtures/test-get-profile-v0.json"

	if _, err := os.Stat(profilePath); err != nil {
		t.Fatal(err)
	}

	outManifest, err := getProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(inManifest, outManifest) {
		t.Fatalf("manifests do not match with each other.\nin:%v\nout:%v\n", inManifest, outManifest)
	}
}

func TestPutGetProfile(t *testing.T) {
	tmp, err := ioutil.TempFile(os.TempDir(), "test-put-profile.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	profilePath := tmp.Name()
	if err := putProfile(profilePath, os.FileMode(0755), inManifest); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(profilePath); err != nil {
		t.Fatal(err)
	}

	outManifest, err := getProfile(profilePath)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(inManifest, outManifest) {
		t.Fatalf("manifests do not match with each other.\nin:%v\nout:%v\n", inManifest, outManifest)
	}
}
