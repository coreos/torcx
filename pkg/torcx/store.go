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
	"os"
	"path/filepath"
	"strings"

	"archive/tar"
	"compress/gzip"
	"github.com/Sirupsen/logrus"
)

// StoreCache holds a temporary cache for images/references in the store
type StoreCache struct {
	Paths []string
	Refs  map[archive]string
}

// NewStoreCache constructs a new StoreCache using `paths` as lookup directories
func NewStoreCache(paths []string) (StoreCache, error) {
	var cache StoreCache
	cachedRefs := make(map[archive]string, 0)

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
			bundleName := strings.TrimSuffix(name, ".oci.tgz")

			fp, err := os.Open(path)
			if err != nil {
				return nil
			}
			defer fp.Close()

			zrd, err := gzip.NewReader(fp)
			if err != nil {
				return nil
			}

			_ = tar.NewReader(zrd)
			// TODO(lucab): unpack/extract refs out OCI archive
			// (via oci-spec - c/image - umoci)
			dummy := []string{"TODO-dummy-ref"}
			for _, ref := range dummy {
				pkg := archive{
					Image:     bundleName,
					Reference: ref,
				}
				cachedRefs[pkg] = path

				logrus.WithFields(logrus.Fields{
					"name":      bundleName,
					"reference": ref,
					"path":      path,
				}).Debug("new bundle/reference added to cache")
			}
		}

		return nil
	}

	for _, root := range paths {
		_ = filepath.Walk(root, walkFn)
	}

	cache = StoreCache{
		Paths: paths,
		Refs:  cachedRefs,
	}
	return cache, nil
}

// LookupReference looks for a reference in the store, returning the path
// to the bundle containing it
func (sc *StoreCache) LookupReference(pkg archive) (string, error) {

	path, ok := sc.Refs[pkg]
	if ok {
		return path, nil
	}

	return "", fmt.Errorf("reference %q not found", pkg)
}
