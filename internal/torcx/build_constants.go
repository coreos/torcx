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

// These are torcx constants that can be overridden via link arguments at build-time.
const (
	// SealPath is the path where metadata are written once the system has been sealed.
	SealPath = "/run/metadata/torcx"
	// VendorUsrDir is the standard mountpoint for vendor USR partition.
	VendorUsrDir = "/usr"
	// OemDir contains (mutable) assets provided by the oem.
	OemDir = "/usr/share/oem/torcx/"
	// DefaultTagRef is the default image reference looked up in archives.
	DefaultTagRef = "com.coreos.cl"
	// VendorProfileName is the default vendor profile used.
	VendorProfileName = "vendor"
	// OemProfileName is the default oem profile used.
	OemProfileName = "oem"
)
