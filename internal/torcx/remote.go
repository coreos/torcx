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

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/euank/gotmpl"
	"github.com/pkg/errors"
)

var (
	errNilRemote          = errors.New("nil Remote")
	errEmptyUsrMountpoint = errors.New("empty USR mountpoint")
	errEmptyTemplateURL   = errors.New("empty remote URL template")
)

// evaluateURL evaluates the URL template for a remote
// and performs variables substitution sourcing values from
// `/etc/os-release`.
func (r *Remote) evaluateURL(usrMountpoint string) (string, error) {
	if r == nil {
		return "", errNilRemote
	}
	if usrMountpoint == "" {
		return "", errEmptyUsrMountpoint
	}
	if r.TemplateURL == "" {
		return "", errEmptyTemplateURL
	}

	if !needSubstitution(r.TemplateURL) {
		return r.TemplateURL, nil
	}

	osReleasePath := VendorOsReleasePath(usrMountpoint)
	fp, err := os.Open(osReleasePath)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open %s", osReleasePath)
	}
	defer fp.Close()
	osMeta, err := parseOsRelease(fp)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse %s", osReleasePath)
	}
	osMeta["COREOS_USR"] = usrMountpoint

	templateVars := map[string]string{
		"COREOS_BOARD": osMeta["COREOS_BOARD"],
		"COREOS_USR":   osMeta["COREOS_USR"],
		"ID":           osMeta["ID"],
		"VERSION_ID":   osMeta["VERSION_ID"],
	}
	return gotmpl.TemplateString(r.TemplateURL, gotmpl.MapLookup(templateVars))
}

// needSubstitution checks whether a URL template contains any
// variables that need to be evaluated.
func needSubstitution(template string) bool {
	emptyMap := gotmpl.MapLookup(map[string]string{})
	_, err := gotmpl.TemplateString(template, emptyMap)
	needSub := err != nil
	return needSub
}

// parseOsRelease is the parser for os-release.
func parseOsRelease(rd io.Reader) (map[string]string, error) {
	meta := map[string]string{}

	sc := bufio.NewScanner(rd)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == "" {
			continue
		}
		value := strings.Trim(parts[1], `"`)
		if value == "" {
			continue
		}
		meta[parts[0]] = value
	}
	if sc.Err() != nil {
		return meta, sc.Err()
	}
	return meta, nil
}
