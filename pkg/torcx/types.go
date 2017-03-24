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

const (
	// FUSE_PATH is the hardcoded fuse location
	FUSE_PATH = "/run/metadata/torcx"
	// FUSE_PROFILE_NAME is the key label for fuse profile name
	FUSE_PROFILE_NAME = "TORCX_PROFILE_NAME"
	// FUSE_PROFILE_PATH is the key label for fuse profile path
	FUSE_PROFILE_PATH = "TORCX_PROFILE_PATH"
	// FUSE_BINDIR is the key label for fuse bindir
	FUSE_BINDIR = "TORCX_BINDIR"
	// VENDOR_DIR
	VENDOR_DIR = "/usr/share/torcx"
	// ProfileManifestV0K - profile manifest kind, v0
	ProfileManifestV0K = "profile-manifest-v0"
)

// CommonConfig contains runtime configuration items common to all
// torcx subcommands
type CommonConfig struct {
	BaseDir    string
	RunDir     string
	ConfDir    string
	StorePaths []string
}

// ApplyConfig contains runtime configuration items specific to
// the `apply` subcommand
type ApplyConfig struct {
	CommonConfig
	Profile string
}

// ProfileConfig contains runtime configuration items specific to
// the `profile` subcommand
type ProfileConfig struct {
	CommonConfig
	CurrentProfileName string
	CurrentProfilePath string
	NextProfile        string
}

// ProfileManifestV0 holds JSON profile manifest
type ProfileManifestV0 struct {
	Kind  string   `json:"kind"`
	Value Archives `json:"value"`
}
