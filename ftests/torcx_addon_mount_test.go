// Copyright 2018 CoreOS Inc.
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

package ftests

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestExtractsTgzAddon(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: EmptyImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	testAddonExtracted(t, []string{"addon:01.torcx.tgz"}, "addon", "01")
}

func TestExtractsSquashfsAddon(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: EmptyImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	testAddonExtracted(t, []string{"addon:02.torcx.squashfs"}, "addon", "02")
}

func testAddonExtracted(t *testing.T, storeFiles []string, name string, ref string) {
	if err := os.MkdirAll("/var/lib/torcx/store", 0755); err != nil {
		t.Fatalf("could not make torcx store: %v", err)
	}
	for _, file := range storeFiles {
		addToStore(t, file)
	}

	if err := os.MkdirAll("/etc/torcx/profiles/", 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("/etc/torcx/profiles/test.json", []byte(`
{
  "kind": "profile-manifest-v0",
  "value": {
    "images": [
      {
        "name": "`+name+`",
        "reference": "`+ref+`"
      }
    ]
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile("/etc/torcx/next-profile", []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("torcx-generator")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(string(bytes))
	}
	expectDirectoriesMatch(t, "/ftests/testdata/rootfs", "/run/torcx/unpack/"+name)
}

func expectDirectoriesMatch(t *testing.T, lhs, rhs string) {
	lhsPaths := []string{}
	filepath.Walk(lhs, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("walk error on %v: %v", path, err)
		}
		lhsPaths = append(lhsPaths, strings.TrimPrefix(path, lhs))
		return nil
	})
	rhsPaths := make([]string, 0, len(lhsPaths))
	filepath.Walk(rhs, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			t.Fatalf("walk error on %v: %v", path, err)
		}
		rhsPaths = append(rhsPaths, strings.TrimPrefix(path, rhs))
		return nil
	})

	if len(lhsPaths) != len(rhsPaths) {
		t.Fatalf("expected %+v == %+v; mismatched length", lhsPaths, rhsPaths)
	}

	if !reflect.DeepEqual(lhsPaths, rhsPaths) {
		t.Fatalf("expected %+v == %+v", lhsPaths, rhsPaths)
	}
}

func addToStore(t *testing.T, filename string) {
	f, err := os.Open("/ftests/testdata/" + filename)
	if err != nil {
		t.Fatalf("unable to open file %q to copy: %v", filename, err)
	}
	defer f.Close()

	dest, err := os.OpenFile("/var/lib/torcx/store/"+filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("unable to store file %q to write: %v", filename, err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, f); err != nil {
		t.Fatal(err)
	}
}
