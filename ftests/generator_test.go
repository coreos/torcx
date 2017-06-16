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
	"testing"

	"github.com/coreos/torcx/pkg/torcx"
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

	var profManifest torcx.ProfileManifestV0
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
		t.Errorf("Expected %d images, got %d", 1, imgLen)
	}

	expName := "empty_vendor"
	expRef := "com.coreos.cl"
	imgName := profManifest.Value.Images[0].Name
	imgRef := profManifest.Value.Images[0].Reference
	if imgName != expName {
		t.Errorf("Expected image name %q, got %q", expName, imgName)
	}
	if imgRef != expRef {
		t.Errorf("Expected image reference %q, got %q", expRef, imgRef)
	}
}
