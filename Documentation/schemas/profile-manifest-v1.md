# Profile Manifest - v1

A "profile manifest" is a JSON data structure consumed by torcx and usually provided by an external party (e.g. an user) as a configuration file with `.json` extension.
It contains an ordered list of images to a be applied on a system.

## Changes in v1

Profile manifest v1 includes a new field:
 * `remote` to optionally specify a remote providing this addon

## Schema

- kind (string, required)
- value (object, required)
  - images (array, required, fixed-type, not-nil, min-lenght=0)
    - (object)
      - name (string, required)
      - reference (string, required)
      - remote (string, optional)

## Entries

- kind: hardcoded to `profile-manifest-v1` for this schema revision.
  The type+version of this JSON manifest.
- value: object containing a single typed key-value.
  Manifest content.
- value/images: array of single-type objects, arbitrary length.
  List of packages to be unpacked and set up.
- value/images/#: anonymous array entry, object
- value/images/#/name: string, compatible with OCI image name specs.
  Name of the image to unpack.
- value/images/#/reference: string, compatible with OCI image reference specs.
  Referenced image will be locally looked up as a file named
  `${name}:${reference}.torcx.${format}` where `format` may be either `tgz` or
  `squashfs`. If both exist, either may arbitrarily be chosen.
- value/images/#/remote: string.
  Identifier for the remote where this image can be found.

## JSON schema

```json

{
  "$schema": "http://json-schema.org/draft-05/schema#",
  "type": "object",
  "properties": {
    "kind": {
      "type": "string",
      "enum": ["profile-manifest-v1"]
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
              "reference": {
                "type": "string"
              },
              "remote": {
                "type": "string"
              }
            },
            "required": [
              "name",
              "reference"
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
