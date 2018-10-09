# Torcx Remote Contents - v1

A "remote contents" manifest is a JSON data structure hosted by remote repositories providing additional torcx image archives.
It lists a set of images (with specific versions and formats) which can be retrieved from the specific remote.

## Changes in v1

This is loosely based on the tectonic-torcx package list v0, which does not have a formal specification (but looks like [this](https://tectonic-torcx.release.core-os.net/manifests/amd64-usr/1520.5.0/torcx_manifest.json)).

Notable changes in v1 are:
 * JSON schema specification
 * manifest filename is now `torcx_remote_contents.json.asc`
 * JSON content is now wrapped with an OpenPGP armored signature
 * more consistent use of "image" terminology
 * `kind` is no `torcx-remote-contents-v1`
 * `location` takes a relative path which then resolves to `${base_url}/${location}`, or an absolute URL.

## Manifest location and signature

A remote contents manifest for a specific remote can be located by looking for a file called `torcx_remote_contents.json.asc` under the resolved remote `${base_url}`.

This file contains a JSON object (schema described below), wrapped in an OpenPGP armored clearsign signature.

Such signature can be verified against any of the keys specified in the remote manifest.

## Schema

The new schema is as follows:
- kind: `torcx-remote-contents-v1`
- value
  - images (array, required, fixed-type, not-nil, min-lenght=0)
    - name (string required)
    - defaultVersion (string, optional)
    - versions (array, required)
        - format (string, required)
        - hash (string, required)
        - location (string, required)
        - version (string, required)

*NOTE*: `defaultVersion` is used to resolve the default vendor reference/symlink (e.g. `com.coreos.cl`).

## Entries

- kind: hardcoded to `torcx-remote-contents-v1` for this schema revision.
  The type+version of this JSON manifest.
- value: object containing a single typed key-value.
  Manifest content.
- value/images: array of single-type objects, arbitrary length.
  List of images.
- value/images/#: anonymous array entry, object
- value/images/#/name: string.
  Name of the image.
- value/images/#/defaultVersion: string.
  Default version which can be aliased by the default vendor reference (e.g. `com.coreos.cl`).
- value/images/#/versions: array of single-type objects, arbitrary length.
  List of archives.
- value/images/#/versions/#: anonymous array entry, object
- value/images/#/versions/#/format: string.
  Archive format. Allowed values: "tgz", "squashfs".
- value/images/#/versions/#/hash: string.
- value/images/#/versions/#/location: string.
  A relative path which then resolves to `${base_url}/${remoteFile}`, or an absolute URL.
- value/images/#/versions/#/version: string.
  Image version.

## JSON schema

```json

{
  "$schema": "http://json-schema.org/draft-05/schema#",
  "type": "object",
  "properties": {
    "kind": {
      "type": "string",
      "enum": ["torcx-remote-contents-v1"]
    },
    "value": {
      "type": "object",
      "properties": {
        "images": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "name": {
                "type": "string"
              },
              "defaultVersion": {
                "type": "string"
              },
              "versions": {
                "type": "array",
                "items": {
                  "type": "object",
                  "properties": {
                    "format": {
                      "type": "string"
                    },
                    "hash": {
                      "type": "string"
                    },
                    "location": {
                      "type": "string"
                    },
                    "version": {
                      "type": "string"
                    }
                  },
                  "required": [
                    "format",
                    "hash",
                    "location",
                    "version"
                  ]
                }
              }
            },
            "required": [
              "name",
              "versions"
            ]
          }
        }
      },
      "required": [
        "images"
      ]
    }
  },
  "required": [
    "kind",
    "value"
  ]
}

```
