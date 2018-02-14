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
	"path/filepath"
)

const (
	// DefaultRunDir is the default path where torcx unpacks/propagates all runtime assets.
	DefaultRunDir = "/run/torcx/"
	// DefaultBaseDir is the default torcx base directory
	DefaultBaseDir = "/var/lib/torcx/"
	// DefaultConfDir is the default torcx config directory
	DefaultConfDir = "/etc/torcx/"

	// VendorStoreDir is the vendor store path
	VendorStoreDir = VendorDir + "store/"
	// VendorProfilesDir is the vendor profiles path
	VendorProfilesDir = VendorDir + "profiles/"

	// OemStoreDir is the vendor store path
	OemStoreDir = OemDir + "store/"
	// OemProfilesDir is the vendor profiles path
	OemProfilesDir = OemDir + "profiles/"

	// defaultCfgPath is the default path for common torcx config
	defaultCfgPath = DefaultConfDir + "config.json"
)

// InternalUnpackDir is the directory where root filesystems are unpacked.
func (cc *CommonConfig) InternalUnpackDir() string {
	return filepath.Join(cc.BaseDir, "unpack")
}

// UnpackDir is the directory where root filesystems are available at runtime.
// UnpackDir acts as the public interface to the UnpackDir. It is where
// users of torcx may expect to find image's contents.
func (cc *CommonConfig) UnpackDir() string {
	return filepath.Join(cc.RunDir, "unpack")
}

// RunBinDir is the directory where binaries are symlinked.
func (cc *CommonConfig) RunBinDir() string {
	return filepath.Join(cc.RunDir, "bin")
}

// ProfileDirs are the list of directories where we look for profiles.
func (cc *CommonConfig) ProfileDirs() []string {
	return []string{
		VendorProfilesDir,
		OemProfilesDir,
		cc.UserProfileDir(),
	}
}

// RunProfile is the file where we copy the contents of the applied profile.
func (cc *CommonConfig) RunProfile() string {
	return filepath.Join(cc.RunDir, "profile.json")
}

// UserStorePath is the path where user-fetched archives are written.
// An optional target version can be specified for versioned user store.
func (cc *CommonConfig) UserStorePath(version string) string {
	storePath := filepath.Join(cc.BaseDir, "store")
	if version != "" {
		storePath = filepath.Join(storePath, version)
	}
	return storePath
}

// UserProfileDir is where user profiles are written.
func (cc *CommonConfig) UserProfileDir() string {
	return filepath.Join(cc.ConfDir, "profiles")
}

// NextProfile is the path for the `next-profile` selector configuration file.
func (cc *CommonConfig) NextProfile() string {
	return filepath.Join(cc.ConfDir, "next-profile")
}

// ArchiveFilename is the filename (no directory) for the archive of an image.
func (im *Image) ArchiveFilename() string {
	return fmt.Sprintf("%s:%s.torcx.tgz", im.Name, im.Reference)
}
