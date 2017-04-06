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

package untar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ExtractCfg holds configuration for a tar extractor
type ExtractCfg struct {
	// Link - whether to create hard links
	HardLink bool
	// Symlink - whether to create symlinks
	Symlink bool
	// Chown - whether to set file uid and gid
	Chown bool
	// Chmod - whether to set file mode
	Chmod bool
	// Chtimes - whether to set atime and mtime
	Chtimes bool
	// XattrUser - whether to set user extend attributes
	XattrUser bool
	// XattrPrivileged - attempt to set non-user extend attributes
	XattrPrivileged bool
	// UIDShift - positive increment to uid (requires Chown)
	UIDShift uint
	// GIDShift - positive increment to gid (requires Chown)
	GIDShift uint
}

// Default returns a default configuration for extract operations
func (ec ExtractCfg) Default() ExtractCfg {
	return ExtractCfg{
		HardLink:  true,
		Symlink:   true,
		Chown:     true,
		Chmod:     true,
		Chtimes:   true,
		XattrUser: true,
		UIDShift:  0,
		GIDShift:  0,
	}
}

// ExtractRoot reads tar entries from r until EOF and creates
// filesystem entries rooted in "/".
// It is supposed to be used by a process already chroot'ed into
// target destination.
//
// Behavior changes according to flag, bitwise-or of the
// above constants:
//
// If Chmod is unset, dirs are created with mode 0755 and files wit 0666.
// Both are subject to umask.
//
// Only numerical uid and gid are handled, no shift or name resolution
// is applied.
func ExtractRoot(tr *tar.Reader, cfg ExtractCfg) error {
	targetDir := "/"

	if tr == nil {
		return fmt.Errorf("invalid tar reader")
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		err = extractOne(hdr, tr, targetDir, cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractOne(hdr *tar.Header, r io.Reader, targetDir string, cfg ExtractCfg) error {
	// Clean before joining to remove all .. elements
	path := filepath.Join(targetDir, filepath.Clean(hdr.Name))
	fi := hdr.FileInfo()

	// Extract entry
	switch hdr.Typeflag {
	case tar.TypeReg, tar.TypeRegA:
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, fi.Mode())
		if err != nil {
			return err
		}
		if _, err = io.Copy(f, r); err != nil {
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	case tar.TypeLink:
		if !cfg.HardLink {
			return nil
		}
		// Skip adjusting metadata below for hardlinks
		return os.Link(hdr.Linkname, path)
	case tar.TypeSymlink:
		if cfg.Symlink {
			return nil
		}
		if err := os.Symlink(hdr.Linkname, path); err != nil {
			return err
		}
	case tar.TypeDir:
		if err := os.MkdirAll(path, fi.Mode()); err != nil {
			return err
		}
	case tar.TypeChar:
		dev := makedev(uint(hdr.Devmajor), uint(hdr.Devminor))
		mode := uint32(fi.Mode()) | syscall.S_IFCHR
		if err := syscall.Mknod(path, mode, int(dev)); err != nil {
			return err
		}
	case tar.TypeBlock:
		dev := makedev(uint(hdr.Devmajor), uint(hdr.Devminor))
		mode := uint32(fi.Mode()) | syscall.S_IFBLK
		if err := syscall.Mknod(path, mode, int(dev)); err != nil {
			return err
		}
	case tar.TypeFifo:
		if err := syscall.Mkfifo(path, uint32(fi.Mode())); err != nil {
			return err
		}
	case tar.TypeCont, tar.TypeXHeader, tar.TypeXGlobalHeader:
		return nil
	default:
		return fmt.Errorf("extract: unrecognized type %q: %s", hdr.Typeflag, hdr.Name)

	}

	// Adjust metadata
	if cfg.Chtimes {
		atime, mtime := hdr.AccessTime, hdr.ModTime
		if atime.IsZero() {
			atime = time.Now()
		}
		if mtime.IsZero() {
			mtime = time.Now()
		}
		if err := os.Chtimes(path, atime, mtime); err != nil {
			return err
		}
	}
	if cfg.Chmod {
		mode := os.FileMode(hdr.Mode)
		if err := os.Chmod(path, mode); err != nil {
			return err
		}
	}
	if cfg.Chown {
		origUID, oriGGID := hdr.Uid, hdr.Gid
		uid, gid := origUID+int(cfg.UIDShift), oriGGID+int(cfg.GIDShift)
		if err := os.Lchown(path, uid, gid); err != nil {
			return err
		}
	}
	if cfg.XattrPrivileged || cfg.XattrUser {
		for k, v := range hdr.Xattrs {
			if strings.HasPrefix(k, "user") && !cfg.XattrUser {
				continue
			}
			if !strings.HasPrefix(k, "user") && !cfg.XattrPrivileged {
				continue
			}

			err := syscall.Setxattr(path, k, []byte(v), 0)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func makedev(maj uint, min uint) uint64 {
	return uint64(min&0xff) | (uint64(maj&0xfff) << 8) |
		((uint64(min) & ^uint64(0xff)) << 12) |
		((uint64(maj) & ^uint64(0xfff)) << 32)
}
