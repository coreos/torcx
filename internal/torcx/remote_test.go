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

package torcx

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TODO(lucab): add more positive tests

func TestBasicEvaluateURL(t *testing.T) {
	var r *Remote
	url := "https://example.com/basepath"

	_, err := r.evaluateURL("")
	if err != errNilRemote {
		t.Fatalf("expected %s, got %s", errNilRemote, err)
	}
	r = &Remote{}
	_, err = r.evaluateURL("")
	if err != errEmptyUsrMountpoint {
		t.Fatalf("expected %s, got %s", errEmptyUsrMountpoint, err)
	}
	_, err = r.evaluateURL("/usr")
	if err != errEmptyTemplateURL {
		t.Fatalf("expected %s, got %s", errEmptyTemplateURL, err)
	}
	r.TemplateURL = url
	res, err := r.evaluateURL("/usr")
	if err != nil {
		t.Fatalf("got unexpected error %s", err)
	}
	if res.String() != url {
		t.Fatalf("expected %s, got %s", url, res)
	}
}

func TestEvaluateURLTemplating(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "torcx_remote_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	libdir := filepath.Join(tmpDir, "lib")
	if err := os.MkdirAll(libdir, 0755); err != nil {
		t.Fatal(err)
	}
	osReleasePath := filepath.Join(libdir, "os-release")
	osContent := `
ID="coreos"
VERSION_ID="1680.2.0"
COREOS_BOARD="amd64-usr"
`
	if err := ioutil.WriteFile(osReleasePath, []byte(osContent), 0755); err != nil {
		t.Fatal(err)
	}

	basURL := "https://example.com/baseurl/"
	testCases := []struct {
		template string
		result   string
	}{
		{
			"",
			basURL,
		},
		{
			"${ID}",
			basURL + "coreos",
		},
		{
			"${VERSION_ID}",
			basURL + "1680.2.0",
		},
		{
			"${COREOS_BOARD}",
			basURL + "amd64-usr",
		},
		{
			"${COREOS_USR}",
			basURL + tmpDir,
		},
		{
			"${ID}/${COREOS_BOARD}/${VERSION_ID}",
			basURL + "coreos/amd64-usr/1680.2.0",
		},
	}

	for _, tt := range testCases {
		template := basURL + tt.template
		r := Remote{
			TemplateURL: template,
		}
		res, err := r.evaluateURL(tmpDir)
		if err != nil {
			t.Fatalf("got unexpected error %s", err)
		}
		if res.String() != tt.result {
			t.Fatalf("expected %s, got %s", tt.result, res)
		}
	}
}
