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

package cli

import (
	"os"
	"strings"
	"testing"
)

func TestFillCommon(t *testing.T) {
	vendorStore := "/usr/share/torcx/store/"
	tests := []struct {
		desc string

		isErr   bool
		basedir string
		rundir  string
		confdir string
	}{
		{
			"basic",
			false,
			"/var/lib/torcx/",
			"/run/torcx/",
			"/etc/torcx/",
		},
	}

	for _, tt := range tests {
		t.Logf("Testing %q", tt.desc)
		cfg, err := fillCommonRuntime()
		if tt.isErr {
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if cfg != nil {
				t.Fatalf("expected nil config, got %#v", cfg)
			}
			continue
		}
		if cfg == nil {
			t.Fatal("expected config, got nil")
		}
		if cfg.BaseDir != tt.basedir {
			t.Fatalf("wrong basedir: expected %q, got %q", tt.basedir, cfg.BaseDir)
		}
		if cfg.RunDir != tt.rundir {
			t.Fatalf("wrong rundir: expected %q, got %q", tt.rundir, cfg.RunDir)
		}
		if cfg.ConfDir != tt.confdir {
			t.Fatalf("wrong rundir: expected %q, got %q", tt.confdir, cfg.ConfDir)
		}
		if len(cfg.StorePaths) == 0 {
			t.Fatal("no store paths")
		}
		foundVendor := false
		for _, path := range cfg.StorePaths {
			if path == vendorStore {
				foundVendor = true
			}
		}
		if !foundVendor {
			t.Fatalf("vendor store %q not found in %#v", vendorStore, cfg.StorePaths)
		}
	}
}

func TestHasExpFeature(t *testing.T) {
	tests := map[string]bool{
		"a": true,
		"B": true,
		"c": false,
		"A": false,
	}

	for key, expFeat := range tests {
		envKey := "TORCX_EXP_" + strings.ToUpper(key)
		os.Unsetenv(envKey)
		if expFeat {
			os.Setenv(envKey, "y")
		}

		gotFeat := hasExpFeature(key)
		if gotFeat != expFeat {
			t.Errorf("Testcase %q failed, expected %t got %t", key, expFeat, gotFeat)

		}
	}
}
