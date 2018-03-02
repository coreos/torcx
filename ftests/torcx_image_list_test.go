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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/coreos/torcx/internal/cli"
	"github.com/coreos/torcx/internal/torcx"
)

func TestImageListEmpty(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: EmptyImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	cmd := exec.Command("torcx", "image", "list", "-v=error")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(b))
		t.Error(err)
	}

	var imgList cli.ImageList
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&imgList)
	if err != nil {
		t.Log(string(b))
		t.Error(err)
	}

	imgLen := len(imgList.Value)
	if imgLen != 0 {
		t.Log(string(b))
		t.Fatalf("Expected %d images, got %d", 1, imgLen)
	}
}

func TestImageListVendor(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: VendorImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	cmd := exec.Command("torcx", "image", "list", "-v=error")
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(b))
		t.Error(err)
	}

	var imgList cli.ImageList
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&imgList)
	if err != nil {
		t.Log(string(b))
		t.Error(err)
	}

	imgLen := len(imgList.Value)
	if imgLen != 1 {
		t.Log(string(b))
		t.Fatalf("Expected %d images, got %d", 1, imgLen)
	}

	expName := "empty_vendor"
	expRef := "com.coreos.cl"
	expPath := fmt.Sprintf("/usr/share/torcx/store/%s:%s.torcx.tgz", expName, expRef)
	imgName := imgList.Value[0].Name
	imgRef := imgList.Value[0].Reference
	imgPath := imgList.Value[0].Filepath
	if imgName != expName {
		t.Errorf("Expected image name %q, got %q", expName, imgName)
	}
	if imgRef != expRef {
		t.Errorf("Expected image reference %q, got %q", expRef, imgRef)
	}
	if imgPath != expPath {
		t.Errorf("Expected image path %q, got %q", expPath, imgPath)
	}
}

func TestImageListUser(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: VendorImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	expName := "empty_vendor"
	expRef := "com.coreos.cl"
	OSVersion := "1.2.3"
	OSEntry := bytes.NewBufferString(fmt.Sprintf("VERSION_ID=%s\n", OSVersion))
	userStore := "/var/lib/torcx/store"

	if err := os.MkdirAll(filepath.Dir(torcx.OsReleasePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(torcx.OsReleasePath, OSEntry.Bytes(), 0755); err != nil {
		t.Fatal(err)
	}

	// These just re-use the vendor package embedded in the ACI,
	// moving/symlinking it around across stores.
	tests := []struct {
		desc          string
		store         string
		oldPath       string
		imgPath       string
		versionFilter string
		doMove        bool
	}{
		{
			"user store",
			userStore,
			fmt.Sprintf("/usr/share/torcx/store/%s:%s.torcx.tgz", expName, expRef),
			fmt.Sprintf("%s/%s:%s.torcx.tgz", userStore, expName, expRef),
			"",
			true,
		},
		{
			"user versioned store",
			filepath.Join(userStore, OSVersion),
			fmt.Sprintf("%s/%s:%s.torcx.tgz", userStore, expName, expRef),
			fmt.Sprintf("%s/%s/%s:%s.torcx.tgz", userStore, OSVersion, expName, expRef),
			OSVersion,
			false, // Just symlink, and check for proper shadowing
		},
	}

	for _, tt := range tests {
		t.Logf("Testing %q", tt.desc)
		err := os.MkdirAll(tt.store, 0755)
		if err != nil {
			t.Fatal(err)
		}
		if tt.doMove {
			err = os.Rename(tt.oldPath, tt.imgPath)
		} else {
			err = os.Symlink(tt.oldPath, tt.imgPath)
		}
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"image", "list", "-v=error"}
		if tt.versionFilter != "" {
			args = append(args, []string{"-n", tt.versionFilter}...)
		}
		cmd := exec.Command("torcx", args...)
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Log(string(b))
			t.Fatal(err)
		}

		var imgList cli.ImageList
		err = json.NewDecoder(bytes.NewReader(b)).Decode(&imgList)
		if err != nil {
			t.Log(string(b))
			t.Fatal(err)
		}

		imgLen := len(imgList.Value)
		if imgLen != 1 {
			t.Log(string(b))
			t.Fatalf("Expected %d images, got %d", 1, imgLen)
		}
		checkImage(t, imgList.Value[0], expName, expRef, tt.imgPath)
	}

}

func TestImageListVersionFilter(t *testing.T) {
	if !IsInContainer() {
		cfg := RktConfig{
			imageName: VendorImage,
		}
		RunTestInContainer(t, cfg)
		return
	}

	OSVersion := "1.2.3"
	OSEntry := bytes.NewBufferString(fmt.Sprintf("VERSION_ID=%s\n", OSVersion))
	if err := os.MkdirAll(filepath.Dir(torcx.OsReleasePath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(torcx.OsReleasePath, OSEntry.Bytes(), 0755); err != nil {
		t.Fatal(err)
	}

	origName := "empty_vendor"
	origRef := "com.coreos.cl"
	origPath := fmt.Sprintf("/usr/share/torcx/store/%s:%s.torcx.tgz", origName, origRef)
	userStore := "/var/lib/torcx/store"
	versionedStore := filepath.Join(userStore, OSVersion)
	newName := "fake_empty"
	versionedRef := "versioned"
	unversionedRef := "unversioned"
	unversionedPath := fmt.Sprintf("%s/%s:%s.torcx.tgz", userStore, newName, unversionedRef)
	versionedPath := fmt.Sprintf("%s/%s:%s.torcx.tgz", versionedStore, newName, versionedRef)
	err := os.MkdirAll(versionedStore, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(origPath, versionedPath)
	if err != nil {
		t.Fatal(err)
	}
	err = os.Symlink(origPath, unversionedPath)
	if err != nil {
		t.Fatal(err)
	}

	// These just re-use the vendor package embedded in the ACI,
	// moving/symlinking it around across stores.
	tests := []struct {
		desc          string
		versionFilter string
		numImages     int
	}{
		{
			"no version",
			"",
			3,
		},
		{
			"matching version",
			OSVersion,
			3,
		},
		{
			"non-matching version",
			"bad.version",
			1,
		},
	}

	for _, tt := range tests {
		t.Logf("Testing %q", tt.desc)

		args := []string{"image", "list", "-v=error"}
		if tt.versionFilter != "" {
			args = append(args, []string{"-n", tt.versionFilter}...)
		}
		cmd := exec.Command("torcx", args...)
		b, err := cmd.CombinedOutput()
		if err != nil {
			t.Log(string(b))
			t.Fatal(err)
		}

		var imgList cli.ImageList
		err = json.NewDecoder(bytes.NewReader(b)).Decode(&imgList)
		if err != nil {
			t.Log(string(b))
			t.Fatal(err)
		}

		imgLen := len(imgList.Value)
		if imgLen != tt.numImages {
			t.Log(string(b))
			t.Fatalf("Expected %d images, got %d", tt.numImages, imgLen)
		}
	}

}

func checkImage(t *testing.T, entry cli.ImageEntry, expName, expRef, expPath string) {
	imgName := entry.Name
	imgRef := entry.Reference
	imgPath := entry.Filepath
	if imgName != expName {
		t.Errorf("Expected image name %q, got %q", expName, imgName)
	}
	if imgRef != expRef {
		t.Errorf("Expected image reference %q, got %q", expRef, imgRef)
	}
	if imgPath != expPath {
		t.Errorf("Expected image path %q, got %q", expPath, imgPath)
	}
}
