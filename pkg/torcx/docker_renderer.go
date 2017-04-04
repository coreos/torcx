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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	imgcopy "github.com/containers/image/copy"
	imgdocker "github.com/containers/image/docker"
	imgdockerref "github.com/containers/image/docker/reference"
	imgoci "github.com/containers/image/oci/layout"
	imgsig "github.com/containers/image/signature"
)

// DockerFetch fetches a DockerV2 image and stores it as a tgz
// containing the rendered rootfs, returning image details.
func DockerFetch(storeCache StoreCache, storePath, refIn string) (string, error) {
	imageTgz := ""

	remoteRef, err := imgdocker.ParseReference(strings.TrimPrefix(refIn, "docker:"))
	if err != nil {
		return "", err
	}

	// TODO(lucab): update this to take care of digest refs
	components := strings.Split(remoteRef.DockerReference().Name(), "/")
	name := components[len(components)-1]
	remoteTagged, ok := remoteRef.DockerReference().(imgdockerref.NamedTagged)
	if !ok || remoteTagged.Tag() == "latest" {
		remoteTagged, err = imgdockerref.WithTag(remoteRef.DockerReference(), DefaultTagRef)
		if err != nil {
			return "", err
		}
		remoteRef, err = imgdocker.NewReference(remoteTagged)
		if err != nil {
			return "", err
		}
	}
	tag := remoteTagged.Tag()

	im := Image{
		Name:      name,
		Reference: tag,
	}
	if path, ok := storeCache.Images[im]; ok {
		logrus.WithFields(logrus.Fields{
			"name":      im.Name,
			"reference": im.Reference,
			"path":      path,
		}).Warn("Duplicate name/reference found")
		return "", fmt.Errorf(`Skipping "%s:%s", already found at %q`, im.Name, im.Reference, path.Filepath)
	}

	tmpDir, err := ioutil.TempDir("", "torcx_fetch_")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	localRef, err := imgoci.NewReference(tmpDir, remoteTagged.Tag())
	if err != nil {
		return "", err
	}

	// TODO(lucab): update this for DTC / OCI-signatures
	policy, err := imgsig.DefaultPolicy(nil)
	if err != nil {
		return "", err
	}
	policyCtx, err := imgsig.NewPolicyContext(policy)
	if err != nil {
		return "", err
	}
	defer policyCtx.Destroy()

	err = imgcopy.Image(policyCtx, localRef, remoteRef, nil)
	if err != nil {
		return "", err
	}

	imageTgz = filepath.Join(storePath, name+":"+remoteTagged.Tag()+".torcx.tgz")

	fp, err := os.Create(imageTgz)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	gw := gzip.NewWriter(fp)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// TODO(caseyc): this should be rendered to a rootfs instead
	addTar := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == tmpDir {
			return nil
		}

		fiHdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		fiHdr.Name = strings.TrimPrefix(path, tmpDir)
		fiHdr.Name = strings.TrimLeft(fiHdr.Name, "/")

		err = tw.WriteHeader(fiHdr)
		if err != nil {
			return err
		}

		fp, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fp.Close()

		if fi.Mode().IsRegular() {
			_, err = io.Copy(tw, fp)
			if err != nil {
				return err
			}
		}

		return nil
	}

	err = filepath.Walk(tmpDir, addTar)
	if err != nil {
		return "", err
	}

	// TODO(lucab): store metadata in xattr

	return imageTgz, nil
}
