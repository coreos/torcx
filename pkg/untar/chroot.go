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

package untar

import (
	"archive/tar"
	"fmt"
	"os"
	"syscall"
)

// ChrootUntar extracts a tar reader into a target destination, by first
// chroot-ing into it.
func ChrootUntar(tr *tar.Reader, targetDir string, cfg ExtractCfg) error {
	// Note: defers in reverse order to escape the chroot on return
	if tr == nil {
		return fmt.Errorf("invalid tar reader")
	}

	cwdString, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current workdir: %s", err)
	}
	defer os.Chdir(cwdString)

	rootFile, err := os.Open("/")
	if err != nil {
		return fmt.Errorf("error opening current rootdir: %s", err)
	}
	defer rootFile.Close()
	rootFd, err := getDirFd(rootFile)
	if err != nil {
		return err
	}

	err = syscall.Chroot(targetDir)
	if err != nil {
		return fmt.Errorf("failed to chroot to %q: %s", targetDir, err)
	}
	defer syscall.Chroot(".")
	defer syscall.Fchdir(rootFd)

	err = syscall.Chdir("/")
	if err != nil {
		return fmt.Errorf("failed to chdir to rootdir: %v", err)
	}

	err = ExtractRoot(tr, cfg)
	if err != nil {
		return fmt.Errorf("error extracting tar: %v", err)
	}

	return nil
}

// getDirFd checks if dirFile is a directory and returns its fd
func getDirFd(dirFile *os.File) (int, error) {
	dirInfo, err := dirFile.Stat()
	if err != nil {
		return 0, fmt.Errorf("error getting info on %q: %s", dirFile.Name(), err)
	}
	if !dirInfo.IsDir() {
		return 0, fmt.Errorf("%q is not a directory", dirFile.Name())
	}
	return int(dirFile.Fd()), nil
}
