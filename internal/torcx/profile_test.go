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

var inManifest = ProfileManifestV0JSON{
	Kind: "profile-manifest-v0",
	Value: ImagesV0{
		[]ImageV0{
			ImageV0{
				Name:      "test-name",
				Reference: "test-reference",
			},
		},
	},
}

func TestGetProfileV0(t *testing.T) {
	// Schema of profile v0 is described in
	// https://github.com/coreos/torcx/blob/master/Documentation/schemas/profile-manifest-v0.md
	profilePath := "../../fixtures/test-get-profile-v0.json"

	if _, err := os.Stat(profilePath); err != nil {
		t.Fatal(err)
	}

	outManifest, err := getProfileV0(profilePath)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(inManifest, outManifest) {
		t.Fatalf("manifests do not match with each other.\nin:%v\nout:%v\n", inManifest, outManifest)
	}
}

func TestProfileInvalid(t *testing.T) {
	profilePath := "../../fixtures/test-invalid-profile.json"

	if _, err := os.Stat(profilePath); err != nil {
		t.Fatal(err)
	}

	_, err := getProfileV0(profilePath)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestPutGetProfileV0(t *testing.T) {
	tmp, err := ioutil.TempFile("", "test-torcx-putget-profile")
	if err != nil {
		t.Fatal(err)
	}
	profilePath := tmp.Name()
	defer os.Remove(profilePath)

	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	if err := addToProfileV0(profilePath, os.FileMode(0755), inManifest, nil); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(profilePath); err != nil {
		t.Fatal(err)
	}

	outManifest, err := getProfileV0(profilePath)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(inManifest, outManifest) {
		t.Fatalf("manifests do not match with each other.\nin:%v\nout:%v\n", inManifest, outManifest)
	}
}

func TestProfileAddV0(t *testing.T) {
	tmp, err := ioutil.TempFile("", "test-torcx-add-profile")
	if err != nil {
		t.Fatal(err)
	}
	profilePath := tmp.Name()
	defer os.Remove(profilePath)
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}

	testImage := Image{
		Name:      "testName",
		Reference: "testRef",
	}
	if err := AddToProfile(profilePath, testImage); err != nil {
		t.Fatal(err)
	}
	images, err := ReadProfilePath(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(images))
	}
	if !reflect.DeepEqual(testImage, images[0]) {
		t.Fatalf("images do not match with each other.\nin:%v\nout:%v\n", testImage, images[0])
	}
}

func TestMergeImages(t *testing.T) {
	testCases := []struct {
		desc  string
		lower []Image
		upper []Image

		merged []Image
	}{
		{
			"empty total",
			[]Image{},
			[]Image{},

			[]Image{},
		}, {
			"empty lower",
			[]Image{},
			[]Image{
				Image{
					Name:      "foo0",
					Reference: "0",
				},
			},

			[]Image{
				Image{
					Name:      "foo0",
					Reference: "0",
				},
			},
		}, {
			"empty upper",
			[]Image{
				Image{
					Name:      "foo1",
					Reference: "1",
				},
			},
			[]Image{},

			[]Image{
				Image{
					Name:      "foo1",
					Reference: "1",
				},
			},
		}, {
			"ordered merge",
			[]Image{
				Image{
					Name:      "foo1",
					Reference: "1",
				},
			},
			[]Image{
				Image{
					Name:      "foo2",
					Reference: "2",
				},
			},

			[]Image{
				Image{
					Name:      "foo1",
					Reference: "1",
				},
				Image{
					Name:      "foo2",
					Reference: "2",
				},
			},
		}, {
			"re-order and override reference in lower",
			[]Image{
				Image{
					Name:      "foo2",
					Reference: "3",
				},
				Image{
					Name:      "foo1",
					Reference: "1",
				},
			},
			[]Image{
				Image{
					Name:      "foo2",
					Reference: "2",
				},
			},

			[]Image{
				Image{
					Name:      "foo1",
					Reference: "1",
				},
				Image{
					Name:      "foo2",
					Reference: "2",
				},
			},
		}, {
			"remove image in lower",
			[]Image{
				Image{
					Name:      "foo3",
					Reference: "3",
				},
			},
			[]Image{
				Image{
					Name:      "foo3",
					Reference: "",
				},
			},

			[]Image{},
		},
	}

	for _, tt := range testCases {
		result := mergeImages(tt.lower, tt.upper)

		if !reflect.DeepEqual(result, tt.merged) {
			t.Errorf("testcase %q failed: got %#v, expected %#v", tt.desc, result, tt.merged)
		}

	}
}
