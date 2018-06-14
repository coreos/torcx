# Torcx remotes

This design document contains technical details on **torcx remotes**, including remote layout, local configuration and UI.
It is largely based on https://github.com/coreos/bugs/issues/2215.

## Overview

A "_torcx remote_" is a collection of addon images for torcx, served from a remote source, which can be fetched by a Container Linux (CL) node for use by `torcx-generator`.
Fetching of images from a remote should happen at first-boot time (e.g. in initramfs), during each update (i.e. before marking a node as "update installed, reboot needed"), and whenever a user chooses to modify their torcx profile at runtime.

## Source configuration

Remotes are configured on a CL node via specific configuration files in JSON format.
Remotes are typically named in reverse-dotted notation, and their manifest is called `remote.json`.

Remote configuration files can be stored in different paths, listed in order of priority:
 * `/etc/torcx/remotes/<name>/remote.json`
 * `/usr/share/oem/torcx/remotes/<name>/remote.json`
 * `/usr/share/torcx/remotes/<name>/remote.json`

Example: `/etc/torcx/remotes/com.example.foo/remote.json`

*Note*: torcx does not currently source "runtime paths" (under `/run`), unlike systemd.

### CoreOS reserved remotes

CoreOS will initially reserve two remotes for its own usage:
 * `com.coreos.cl.addons`
 * `com.coreos.cl.usr`

The `.usr` remote is locally available via USR-A/USR-B partitions.
The `.addons` remote provides additional images (e.g. older docker versions) via a public bucket.


### Configuration schema

The JSON configuration for a remote, locally stored on each node, is structured as follows:

Schema:
-   kind (string, required)
-   value (object, required)
    -   base\_url (string, required)
    -   keys (array, required, fixed-type, not-nil) - (object)
        -   armored_keyring (string)

Entries:
-   `kind`: hardcoded to `remote-manifest-v0` for this schema revision. The type+version of this JSON manifest.
-   `value`: object containing a single typed key-value. Manifest content.
-   `value/base_url`: template with base URL for the remote. Supported protocols: "http", "https".
-   `value/keys/#`: array of single-type objects, arbitrary length. It contains trusted keys for signature verification.
-   `value/keys/#/armored_keyring`: path to an ASCII-armored OpenPGP keyring, relative to `base_url`.

URL template is evaluated for simple variable substitution. Interpolated variables are:
 * `${COREOS_BOARD}`: board type (e.g. "amd64-usr")
 * `${COREOS_USR}`: path to a USR mountpoint (e.g. "/usr")
 * `${VERSION_ID}`: OS version (e.g. "1680.2.0")
 * `${ID}`: OS vendor ID (e.g. "coreos")

Example:
```json
{
  "kind":"remote-manifest-v0",  
  "value":{
    "base_url":"https://example.com/torcx-repo/{{.Board}}/{{.OSVersion}}/",  
    "keys":[{
      "armored_keyring": "0xAAAAAABBBBBB.asc"
    }]
  }
}
```
*Note*: the `${COREOS_USR}` entry is used to locate the temporary mountpoint used by `update-engine`.

### Changes to profile manifest

[Profile manifest](https://github.com/coreos/torcx/blob/master/Documentation/schemas/profile-manifest-v0.md) needs to be augmented with a reference to a specific remote for each image.

While this can be done without breaking the current schema, it is preferred to introduce a new "v1" revision so that consuming software can better handle the transition.
The new schema is as follow:
- kind: `profile-manifest-v1`
- value
-   images
    -   image
    -   reference
    -   remote (string, requires, non-empty): name of a locally configured remote.

Example:
```json
{
  "kind":"profile-manifest-v1",
  "value":[{
      "image":"docker",
      "reference":"18.02",
      "remote":"com.coreos.usr"
  }]
}
```

## Remote layout

Remotes are simple collections of addon images served over HTTP(S), together with a signed manifest.
Remotes are keyed by their unique reverse-dotted name and configured locally on each node.

Each remote has a specific base URL, which can be templated and locally reified at runtime by the node.

### Serving a remote

The entrypoint to a remote is its signed manifest. From a configured node, it can be located as follows:
 `manifest_url: ${base_url}/torcx_manifest.json.asc`

Where `base_url` is the reified base URL of the remote, and `torcx_manifest.json.asc` is a fixed-name file containing an armor-signed JSON manifest of remote contents.

A remote will typically contain multiple manifests, one per each supported OS board+version combination.

### Contents manifest

This is loosely based on the tectonic-torcx package list, which currently looks like https://tectonic-torcx.release.core-os.net/manifests/amd64-usr/1520.5.0/torcx_manifest.json

Notable changes are:
 * manifest filename is now `torcx_manifest.json.asc`
 * JSON content is now wrapped with an OpenPGP armored signature
 * s/packages/images/ (i.e. avoid "package" terminology, and align with other manifests)
 * `kind`: `torcx-remote-contents-v1`
 * `location` : takes a relative path which then resolves to `${base_url}/${location}`, or an absolute URL.

The new schema is as follows:
- kind: `torcx-remote-contents-v1`
- value
  - images
    -   name (string required)
    -   defaultVersion (string, optional)
    -   versions (array, required)
        -   version (string, required)
        -   hash (string, required)
        -   sourcePackage (string, optional)
        -   location (string, required)
        -   format (string, required)

*NOTE*: `defaultVersion` is used to resolve the `com.coreos.cl` version symlink.

## torcx UI changes

Torcx will grow some new subcommands to help integrating remotes and `update_engine`.

### torcx profile populate

`torcx profile populate` will be used by update\_engine postinst to populate current and next store, retrieving all images required to satisfy a profile.

Profile fetching is a no-op on images without a remote.

It will check for an additional environment variable:
 * `${TORCX_USR_MOUNTPOINT}`: mountpoint for USR (default: `/usr`)

This command by default operates on the active USR partition. `coreos_postinst` however needs to perform setup for the next OS release, whose USR mountpoint is under a temporary directory.

### torcx profile check

`torcx profile populate` already exists, it will be augmented to check for two additional options:
 * `${TORCX_USR_MOUNTPOINT}`: mountpoint for USR (default: `/usr`)
 * `--skip-remoteless`: skip checking for remote-less images (default: `false`) 

This command currently fails on remote-less images. However, `coreos-postinst` will set `--skip-remoteless=true` in order to keep compatibility with existing profiles (e.g.  the `tectonic` one).

## update\_engine changes

update\_engine postinst will call `torcx profile check` and `torcx profile populate` to ensure the store contains images to satisfy the configured profile on both current and next USR.

Update will fail if a final `torcx profile check` returns with error.


## Bootengine changes

Bootengine will gain a new service to be run from the initramfs: `torcx-profile-populate.service`.

This service will run only on the first-boot. It runs after the ignition `files` stage and before `pivot_root`, requiring network to be up.
It will chroot into `/sysroot` and performs a `torcx profile populate` in there.

This service will needs to succeed if a torcx profile is configured and non-empty, otherwise node provisioning will fail.


