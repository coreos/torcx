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
	// FUSE_UNPACKDIR is the key label for fuse unpackdir
	FUSE_UNPACKDIR = "TORCX_UNPACKDIR"
	// VENDOR_DIR
	VENDOR_DIR = "/usr/share/torcx"
	// ProfileManifestV0K - profile manifest kind, v0
	ProfileManifestV0K = "profile-manifest-v0"
	// ImageManifestV0K - image manifest kind, v0
	ImageManifestV0K = "image-manifest-v0"
	// DefaultTagRef is the default image reference looked up in archives
	DefaultTagRef = "com.coreos.cl"
	// DEFAULT_PROFILE_NAME is the default profile name used
	DEFAULT_PROFILE_NAME = "vendor"
	// CommonConfigV0K - common torcx config kind, v0
	CommonConfigV0K = "torcx-config-v0"
)

// ConfigV0 holds common torcx configuration in JSON format
type ConfigV0 struct {
	Kind  string       `json:"kind"`
	Value CommonConfig `json:"value"`
}

// CommonConfig contains runtime configuration items common to all
// torcx subcommands
type CommonConfig struct {
	BaseDir    string   `json:"base_dir,omitempty"`
	RunDir     string   `json:"run_dir,omitempty"`
	ConfDir    string   `json:"conf_dir,omitempty"`
	StorePaths []string `json:"store_paths,omitempty"`
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
	Kind  string `json:"kind"`
	Value Images `json:"value"`
}

// Archive represents a .torcx.tgz on disk
type Archive struct {
	Image
	Filepath string `json:"filepath"`
}

// Image is an archive name + Reference
type Image struct {
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

type Images struct {
	Images []Image `json:"images"`
}

// ImageManifestV0 holds JSON image manifest
type ImageManifestV0 struct {
	Kind  string `json:"kind"`
	Value Assets `json:"value"`
}

// Assets holds lists of assets propagated from an image to the system
type Assets struct {
	Binaries []string `json:"bin,omitempty"`
	Services []string `json:"service,omitempty"`
}
