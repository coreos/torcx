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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

const (
	// RktEnvFlag is an environment flag to signal execution inside rkt context
	RktEnvFlag = "RKT_FTESTS"
)

var (
	// LocalImages contains a list of local images to pre-pull
	LocalImages = []string{}
	// RemoteImages contains a list of remote images to pre-pull
	RemoteImages = []string{}
)

// RktConfig contains configuration for the rkt test run
type RktConfig struct {
	imageName string
}

// IsInContainer returns whether it is being executed within rkt context
func IsInContainer() bool {
	return os.Getenv(RktEnvFlag) != ""
}

// RunTestInContainer re-runs current tests inside a rkt container
func RunTestInContainer(t *testing.T, cfg RktConfig) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	rktOpts := []string{
		"--insecure-options=all",
		"--net=none",
		"run",
	}
	args := rktOpts
	testsuiteMounts := []string{
		fmt.Sprintf("--volume=runner,kind=host,source=%s,readOnly=true", exe),
		"--mount=volume=runner,target=/runner",
		fmt.Sprintf("--volume=ftests,kind=host,source=%s,readOnly=true", cwd),
		"--mount=volume=ftests,target=/ftests",
		fmt.Sprintf("--volume=torcx-bin,kind=host,source=%s,readOnly=true", filepath.Join(cwd, "..", "bin", runtime.GOARCH, "torcx")),
		"--mount=volume=torcx-bin,target=/usr/bin/torcx",
		"--mount=volume=torcx-bin,target=/usr/bin/torcx-generator",
	}
	// For the squashfs code to create a loopback mount, it needs access to /dev/loop*
	testsuiteMounts = append(testsuiteMounts, loopbackMounts(t)...)
	args = append(args, testsuiteMounts...)
	if cfg.imageName == "" {
		cfg.imageName = "docker://busybox"
	}
	appOpts := []string{
		cfg.imageName,
		fmt.Sprintf("--environment=%s=%s", RktEnvFlag, "true"),
		"--exec=/runner",
		"--", "-test.v", "--test.run", t.Name(),
	}
	args = append(args, appOpts...)
	cmd := exec.Command("rkt", args...)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(string(bytes))
	}
}

func loopbackMounts(t *testing.T) []string {
	mounts := []string{
		"--volume=loopctl,kind=host,source=/dev/loop-control", "--mount=volume=loopctl,target=/dev/loop-control",
	}

	for i := 0; true; i++ {
		_, err := os.Stat(fmt.Sprintf("/dev/loop%d", i))
		if err != nil && os.IsNotExist(err) {
			return mounts
		}
		if err != nil {
			t.Fatalf("error looking for loop%d device: %v", i, err)
		}
		mounts = append(mounts,
			fmt.Sprintf("--volume=loop%d,kind=host,source=/dev/loop%d", i, i),
			fmt.Sprintf("--mount=volume=loop%d,target=/dev/loop%d", i, i),
		)
	}
	return mounts
}
