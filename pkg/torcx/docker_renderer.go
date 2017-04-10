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
	"compress/gzip"
	"fmt"
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

	imgtools "github.com/squeed/image-tools/image"

	pkgtar "github.com/coreos/torcx/pkg/tar"
)

// DockerFetch fetches a DockerV2 image and stores it as a tgz
// containing the rendered rootfs, returning image details.
func DockerFetch(storeCache StoreCache, storePath, refIn string) (*Archive, error) {
	remoteRef, err := imgdocker.ParseReference(strings.TrimPrefix(refIn, "docker:"))
	if err != nil {
		return nil, err
	}

	// TODO(lucab): update this to take care of digest refs
	components := strings.Split(remoteRef.DockerReference().Name(), "/")
	name := components[len(components)-1]
	remoteTagged, ok := remoteRef.DockerReference().(imgdockerref.NamedTagged)
	if !ok || remoteTagged.Tag() == "latest" {
		remoteTagged, err = imgdockerref.WithTag(remoteRef.DockerReference(), DefaultTagRef)
		if err != nil {
			return nil, err
		}
		remoteRef, err = imgdocker.NewReference(remoteTagged)
		if err != nil {
			return nil, err
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
		}).Warn("Image already exists")
		return nil, fmt.Errorf(`Skipping "%s:%s", already found at %q`, im.Name, im.Reference, path.Filepath)
	}

	tmpDir, err := ioutil.TempDir("", "torcx_fetch_")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmpDir)

	localRef, err := imgoci.NewReference(tmpDir, remoteTagged.Tag())
	if err != nil {
		return nil, err
	}

	// TODO(lucab): update this for DTC / OCI-signatures
	policy, err := imgsig.DefaultPolicy(nil)
	if err != nil {
		return nil, err
	}
	policyCtx, err := imgsig.NewPolicyContext(policy)
	if err != nil {
		return nil, err
	}
	defer policyCtx.Destroy()

	// Actually fetch the image
	err = imgcopy.Image(policyCtx, localRef, remoteRef, nil)
	if err != nil {
		return nil, err
	}

	// Render the image to a rootfs archive
	archiveTgz := filepath.Join(storePath, name+":"+remoteTagged.Tag()+".torcx.tgz")
	if err := ExtractImage(tmpDir, im.Reference, archiveTgz); err != nil {
		return nil, err
	}

	// TODO(lucab): store metadata in xattr

	return &Archive{im, archiveTgz}, nil
}

// ExtractImage renders the OCI image to a root filesystem, then creates an Archive.
func ExtractImage(ociPath, ref, dstFile string) error {

	// First, render to a temporary directory
	unpackDir, err := ioutil.TempDir("", "torcx_unpack_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(unpackDir)

	logrus.WithFields(logrus.Fields{
		"path":      ociPath,
		"unpackDir": unpackDir,
		"reference": ref,
	}).Info("Unpacking oci image")

	imageType, err := imgtools.Autodetect(ociPath)
	if err != nil {
		return err
	}

	switch imageType {
	case imgtools.TypeImageLayout:
		err = imgtools.UnpackLayout(ociPath, unpackDir, ref)
	case imgtools.TypeImage:
		err = imgtools.UnpackFile(ociPath, unpackDir, ref)
	default:
		return fmt.Errorf("Unknown image type %s", imageType)
	}

	// Then, tar+gz up the rendered image
	fp, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	gw := gzip.NewWriter(fp)
	defer gw.Close()

	if err := pkgtar.Create(gw, unpackDir); err != nil {
		return err
	}

	// Capture failures to close - they are real errors
	if err := gw.Close(); err != nil {
		return err
	}
	return fp.Close()
}
