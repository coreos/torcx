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
	"os"
	"os/exec"
	"testing"
)

func TestExperimentTorcxFetch(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: EmptyImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	cmd := exec.Command("torcx", "image", "fetch")

	// Flag unset: help message, exit=0
	os.Unsetenv("TORCX_EXP_FETCH")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(string(bytes))
	}

	// Flag set: missing additional parameters, exit=1
	os.Setenv("TORCX_EXP_FETCH", "true")
	bytes, err = cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
