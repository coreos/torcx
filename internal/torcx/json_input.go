// Copyright 2018 CoreOS Inc.
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
	// ProfileManifestV1K - profile manifest kind, v1
	ProfileManifestV1K = "profile-manifest-v1"
	// ProfileManifestV0K - profile manifest kind, v0
	ProfileManifestV0K = "profile-manifest-v0"
	// RemoteManifestV0K - remote manifest kind, v0
	RemoteManifestV0K = "remote-manifest-v0"
	// RemoteContentsV1K - remote contents kind, v1
	RemoteContentsV1K = "torcx-remote-contents-v1"
)

// * Profile manifest version 1: added "remote".

// ProfileManifestV1JSON holds JSON profile manifest (version 1).
type ProfileManifestV1JSON struct {
	Kind  string   `json:"kind"`
	Value ImagesV1 `json:"value"`
}

// ImagesV1 contains an array of image entries.
type ImagesV1 struct {
	Images []ImageV1 `json:"images"`
}

// ImageV1 describes and addon image within a v1 profile.
type ImageV1 struct {
	Name      string `json:"name"`
	Reference string `json:"reference"`
	Remote    string `json:"remote"`
}

// * Profile manifest version 0: initial version.

// ProfileManifestV0JSON holds JSON profile manifest (version 0).
type ProfileManifestV0JSON struct {
	Kind  string   `json:"kind"`
	Value ImagesV0 `json:"value"`
}

// ImagesV0 contains an array of image entries.
type ImagesV0 struct {
	Images []ImageV0 `json:"images"`
}

// ImageV0 represents an addon archive (name + reference).
type ImageV0 struct {
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

// * Remote manifest version 0: initial version.

// RemoteManifestV0JSON holds a JSON remote manifest (version 0).
type RemoteManifestV0JSON struct {
	Kind  string   `json:"kind"`
	Value RemoteV0 `json:"value"`
}

// RemoteV0 describes a remote.
type RemoteV0 struct {
	BaseURL string        `json:"base_url"`
	Keys    []RemoteKeyV0 `json:"keys"`
}

// RemoteKeyV0 represents a signing key for a remote.
type RemoteKeyV0 struct {
	ArmoredKeyring string `json:"armored_keyring,omitempty"`
}

// * Remote contents version 1: initial version.

// RemoteContentsV1JSON holds JSON contents metadata for a remote manifest.
type RemoteContentsV1JSON struct {
	Kind  string         `json:"kind"`
	Value RemoteImagesV1 `json:"value"`
}

// RemoteImagesV1 lists all images available on a remote.
type RemoteImagesV1 struct {
	Images []RemoteImageV1 `json:"images"`
}

// RemoteImageV1 describes image versions available on a remote.
type RemoteImageV1 struct {
	DefaultVersion string            `json:"defaultVersion"`
	Name           string            `json:"name"`
	Versions       []RemoteVersionV1 `json:"versions"`
}

// RemoteVersionV1 describes a specific image (with version and format)
// available on a remote.
type RemoteVersionV1 struct {
	Format   string `json:"format"`
	Hash     string `json:"hash"`
	Location string `json:"location"`
	Version  string `json:"version"`
}
