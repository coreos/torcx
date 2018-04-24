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

package ftests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/coreos/torcx/internal/torcx"
)

const (
	dockerFlagPath = "/etc/coreos/docker-1.12"
)

func TestGeneratorEmpty(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: EmptyImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	cmd := exec.Command("torcx-generator")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(string(bytes))
	}
}

func TestGeneratorNoNextprofile(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: VendorImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	cmd := exec.Command("torcx-generator")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(string(bytes))
	}

	checkProfileSelection(t, "empty_vendor", "com.coreos.cl")
}

func TestGeneratorDockerFlagDefault(t *testing.T) {
	testDockerFlag(t, nil, "com.coreos.cl")
}

func TestGeneratorDockerFlagYes(t *testing.T) {
	testDockerFlag(t, makeFile(dockerFlagPath, "yes"), "1.12")
}

func TestGeneratorDockerFlagNo(t *testing.T) {
	testDockerFlag(t, makeFile(dockerFlagPath, "no\n"), "17.06")
}

func TestGeneratorDockerFlagBadValue(t *testing.T) {
	testDockerFlag(t, makeFile(dockerFlagPath, "aardvark"), "com.coreos.cl")
}

func TestGeneratorDockerFlagUnreadable(t *testing.T) {
	setup := func(t *testing.T) {
		if err := os.MkdirAll(dockerFlagPath, 0777); err != nil {
			t.Fatal(err)
		}
	}
	testDockerFlag(t, setup, "com.coreos.cl")
}

func TestGeneratorDockerFlagOverride(t *testing.T) {
	setup := func(t *testing.T) {
		makeFile(dockerFlagPath, "no")(t)
		makeFile("/etc/torcx/next-profile", "docker-1.12-yes")(t)
	}
	testDockerFlag(t, setup, "1.12")
}

func TestGeneratorDockerProfileMissing(t *testing.T) {
	setup := func(t *testing.T) {
		makeFile(dockerFlagPath, "yes")(t)
		if err := os.Remove("/usr/share/torcx/profiles/docker-1.12-yes.json"); err != nil {
			t.Fatal(err)
		}
	}
	testDockerFlag(t, setup, "com.coreos.cl")
}

func makeFile(path, contents string) func(*testing.T) {
	return func(t *testing.T) {
		if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if _, err := f.WriteString(contents); err != nil {
			t.Fatal(err)
		}
	}
}

func testDockerFlag(t *testing.T, setup func(*testing.T), expRef string) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: DockerFlagImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	if setup != nil {
		setup(t)
	}

	cmd := exec.Command("torcx-generator")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(string(bytes))
	}

	checkProfileSelection(t, "empty_vendor", expRef)
}

func checkProfileSelection(t *testing.T, expName, expRef string) {
	var profManifest torcx.ProfileManifestV0JSON
	fp, err := os.Open("/run/torcx/profile.json")
	if err != nil {
		t.Error(err)
	}

	err = json.NewDecoder(fp).Decode(&profManifest)
	if err != nil {
		t.Error(err)
	}

	imgLen := len(profManifest.Value.Images)
	if imgLen != 1 {
		t.Fatalf("Expected %d images, got %d", 1, imgLen)
	}

	imgName := profManifest.Value.Images[0].Name
	imgRef := profManifest.Value.Images[0].Reference
	if imgName != expName {
		t.Errorf("Expected image name %q, got %q", expName, imgName)
	}
	if imgRef != expRef {
		t.Errorf("Expected image reference %q, got %q", expRef, imgRef)
	}
}
