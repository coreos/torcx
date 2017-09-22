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
	"testing"
)

const (
	EmptyImage      = "github.com/coreos/torcx/ftests/empty"
	VendorImage     = "github.com/coreos/torcx/ftests/vendor"
	DockerFlagImage = "github.com/coreos/torcx/ftests/docker-flag"
)

func init() {
	LocalImages = append(LocalImages,
		[]string{
			"fixtures/empty.aci",
			"fixtures/vendor.aci",
			"fixtures/docker-flag.aci",
		}...)
}

func TestMain(m *testing.M) {
	exitCode := 0
	if err := testRunner(m); err != nil {
		fmt.Printf("[ERR] %s\n", err)
		exitCode = 1
	}
	os.Exit(exitCode)
}

func testRunner(m *testing.M) error {
	if !IsInContainer() {
		if err := runnerSetUp(); err != nil {
			return err
		}
		defer runnerTearDown()
	}
	if m.Run() != 0 {
		return fmt.Errorf("Functional testsuite failed")
	}
	return nil
}

func runnerSetUp() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("Functional testsuite must run as root/sudo")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	rktBin, err := exec.LookPath("rkt")
	if err != nil {
		return err
	}
	for _, aci := range LocalImages {
		if err := exec.Command(rktBin, "--insecure-options=image", "fetch", filepath.Join(cwd, aci)).Run(); err != nil {
			return err
		}
	}
	return nil
}

func runnerTearDown() error {
	return exec.Command("rkt", "gc", "--grace-period=0s").Run()
}
