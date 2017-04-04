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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
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

	walkFn := func(inPath string, inInfo os.FileInfo, inErr error) error {
		if inErr != nil {
			return nil
		}
		path := filepath.Clean(inPath)
		name := filepath.Base(path)

		if !inInfo.Mode().IsRegular() {
			return nil
		}
		if !strings.HasSuffix(name, ".oci.tgz") {
			return nil
		}

		if inInfo.Mode().IsRegular() {
			archive := Archive{
				Name:     strings.TrimSuffix(name, ".oci.tgz"),
				Filepath: path,
			}

			fp, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer fp.Close()

			zrd, err := gzip.NewReader(fp)
			if err != nil {
				return nil
			}

			trd := tar.NewReader(zrd)
			for {
				hdr, err := trd.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}
				if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
					continue
				}
				// TODO(lucab): this needs a complete overhaul for OCI -rc5
				if !strings.HasPrefix(hdr.Name, "refs/") {
					continue
				}
				// TODO(lucab): add manifest validation

				image := Image{
					Name:      archive.Name,
					Reference: strings.TrimPrefix(hdr.Name, "refs/"),
				}

				// The first archive to define a reference always wins,
				// warn on collision
				if _, ok := sc.Images[image]; ok {
					logrus.WithFields(logrus.Fields{
						"name":      image.Name,
						"reference": image.Reference,
						"path":      path,
					}).Warn("Duplicate name + reference ignored!")
				} else {
					logrus.WithFields(logrus.Fields{
						"name":      image.Name,
						"reference": image.Reference,
						"path":      path,
					}).Debug("new archive/reference added to cache")

					sc.Images[image] = archive
				}
			}
		}

		return nil
	}

	for _, root := range paths {
		_ = filepath.Walk(root, walkFn)
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
