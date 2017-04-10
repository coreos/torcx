// Copyright 2014 The rkt Authors
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
//
// Simplified from https://github.com/rkt/pkg/tar/

package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Create(w io.Writer, root string) error {
	if stat, err := os.Stat(root); err != nil {
		return err
	} else if !stat.IsDir() {
		return fmt.Errorf("tar root must be a directory: %s", root)
	}

	tw := tar.NewWriter(w)

	err := filepath.Walk(root,
		func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			linkDest := ""
			if fi.Mode()&os.ModeSymlink > 0 {
				linkDest, err = os.Readlink(path)
				if err != nil {
					return err
				}
			}

			tarHeader, err := tar.FileInfoHeader(fi, linkDest)
			tarHeader.Name, err = filepath.Rel(root, path)
			if err != nil {
				return err
			}

			if err := tw.WriteHeader(tarHeader); err != nil {
				return err
			}

			// Skip non-files
			if fi.Mode()&os.ModeType > 0 {
				return nil
			}

			fp, err := os.Open(path)
			if err != nil {
				return err
			}
			defer fp.Close()

			_, err = io.Copy(tw, fp)
			if err != nil {
				return err
			}

			return nil
		})
	if err != nil {
		tw.Close()
		return err
	}

	return tw.Close()
}
