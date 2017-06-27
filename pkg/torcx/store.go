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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
)

// StoreCache holds a temporary cache for images/references in the store
type StoreCache struct {
	Paths []string

	// The mapping of name + reference to .tgz file
	Images map[Image]Archive
}

// NewStoreCache constructs a new StoreCache using `paths` as lookup directories
func NewStoreCache(paths []string) (StoreCache, error) {
	sc := StoreCache{
		Paths:  paths,
		Images: map[Image]Archive{},
	}

	walkFn := func(dir string, inInfo os.FileInfo) error {
		path := filepath.Clean(filepath.Join(dir, inInfo.Name()))
		name := filepath.Base(path)

		// Ensure a symlink points to a regular file
		if inInfo.Mode()&os.ModeSymlink != 0 {
			if lpath, err := filepath.EvalSymlinks(path); err != nil {
				return nil
			} else if inInfo, err = os.Lstat(lpath); err != nil {
				return nil
			}
		}

		if !inInfo.Mode().IsRegular() {
			return nil
		}
		if !strings.HasSuffix(name, ".torcx.tgz") {
			return nil
		}
		baseName := strings.TrimSuffix(name, ".torcx.tgz")
		imageName := baseName
		imageRef := DefaultTagRef
		if strings.ContainsRune(baseName, ':') {
			subs := strings.Split(baseName, ":")
			imageRef = subs[len(subs)-1]
			imageName = strings.Join(subs[:len(subs)-1], "")
		}

		image := Image{
			Name:      imageName,
			Reference: imageRef,
		}
		archive := Archive{image, path}

		// The first archive to define a reference always wins,
		// warn on collision
		if ar, ok := sc.Images[image]; ok {
			logrus.WithFields(logrus.Fields{
				"name":      image.Name,
				"reference": image.Reference,
				"original":  ar.Filepath,
				"duplicate": path,
			}).Warn("skipped duplicate image")
		} else {
			logrus.WithFields(logrus.Fields{
				"name":      image.Name,
				"reference": image.Reference,
				"path":      path,
			}).Debug("new archive/reference added to cache")

			sc.Images[image] = archive
		}

		return nil
	}

	for _, dir := range paths {
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"path": dir,
				"err":  err,
			}).Info("store skipped")
			continue
		}
		for _, fi := range files {
			_ = walkFn(dir, fi)
		}
	}

	return sc, nil
}

// LookupReference looks for a reference in the store, returning the path
// to the archive containing it
func (sc *StoreCache) ArchiveFor(im Image) (Archive, error) {

	arch, ok := sc.Images[im]
	if ok {
		return arch, nil
	}

	return Archive{}, fmt.Errorf("image %s:%s not found", im.Name, im.Reference)
}
