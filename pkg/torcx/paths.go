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

// RunUnpackDir is the directory where root filesystems are unpacked.
func (cc *CommonConfig) RunUnpackDir() string {
	return filepath.Join(cc.RunDir, "unpack")
}

// RunBinDir is the directory where binaries are symlinked.
func (cc *CommonConfig) RunBinDir() string {
	return filepath.Join(cc.RunDir, "bin")
}

// ProfileDirs are the list of directories where we look for profiles.
func (cc *CommonConfig) ProfileDirs() []string {
	return []string{
		filepath.Join(VENDOR_DIR, "profiles.d"),
		cc.UserProfileDir(),
	}
}

// RunProfile is the file where we copy the contents of the applied profile.
func (cc *CommonConfig) RunProfile() string {
	return filepath.Join(cc.RunDir, "profile.json")
}

// UserStorePath  is the path where user-fetched archives are written.
func (cc *CommonConfig) UserStorePath() string {
	return filepath.Join(cc.BaseDir, "store")
}

// AuthDir will have docker trust roots. It is currently unused.
func (cc *CommonConfig) AuthDir() string {
	return filepath.Join(cc.ConfDir, "auth.d")
}

// UserProfileDir is where user profiles are written.
func (cc *CommonConfig) UserProfileDir() string {
	return filepath.Join(cc.ConfDir, "profiles.d")
}

// NextProfile is the path for the `next-profile` selector configuration file.
func (cc *CommonConfig) NextProfile() string {
	return filepath.Join(cc.ConfDir, "next-profile")
}

// ArchiveFilename is the filename (no directory) for the archive of an image.
func (im *Image) ArchiveFilename() string {
	return fmt.Sprintf("%s:%s.torcx.tgz", im.Name, im.Reference)
}
