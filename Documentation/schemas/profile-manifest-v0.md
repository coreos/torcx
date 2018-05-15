# Profile Manifest - v0

A "profile manifest" is a JSON data structure consumed by torcx and usually provided by an external party (e.g. an user) as a configuration file with `.json` extension.
It contains an ordered list of images (name + reference) to a be applied on a system.

## Schema

- kind (string, required)
- value (object, required)
  - images (array, required, fixed-type, not-nil, min-lenght=0)
    -(object)
      - image (string, required)
      - reference (string, required)

## Entries

- kind: hardcoded to `profile-manifest-v0` for this schema revision.
  The type+version of this JSON manifest.
- value: object containing a single typed key-value.
  Manifest content.
- value/images: array of single-type objects, arbitrary length.
  List of packages to be unpacked and set up.
- value/images/#: anonymous array entry, object
- value/images/#/image: string, compatible with OCI image name specs.
  Name of the image to unpack.
- value/images/#/reference: string, compatible with OCI image reference specs.
  Referenced image must be available in the storepath, as a file name `${image}:${reference}.torcx.tgz`.

## JSON schema

```json

{
  "$schema": "http://json-schema.org/draft-05/schema#",
  "type": "object",
  "properties": {
    "kind": {
      "type": "string",
      "enum": ["profile-manifest-v0"]
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
