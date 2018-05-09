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
