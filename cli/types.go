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

package cli

const (
	// TorcxProfileListV0 is the JSON kind identifier for a profile list
	TorcxProfileListV0 = "torcx-profile-list-v0"
)

// ProfileList is the JSON container for profile list output
type ProfileList struct {
	Kind  string      `json:"kind"`
	Value profileList `json:"value"`
}

type profileList struct {
	CurrentProfileName *string  `json:"current_profile_name"`
	CurrentProfilePath *string  `json:"current_profile_path"`
	NextProfileName    string   `json:"next_profile_name"`
	Profiles           []string `json:"profiles"`
}
