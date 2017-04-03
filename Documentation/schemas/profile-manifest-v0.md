# Profile Manifest - v0

## Schema

- kind (string, required)
- value (object, required)
  - archives (array, required, fixed-type, not-nil, min-lenght=0)
    -(object)
      - image (string, required)
      - reference (string, required)

## Entries

- kind: hardcoded to `profile-manifest-v0` for this schema revision.
  The type+version of this JSON manifest.
- value: object containing a single typed key-value.
  Manifest content.
- value/archives: array of single-type objects, arbitrary length.
  List of packages to be unpacked and setup.
- value/archives/#: anonymous array entry, object
- value/archives/#/image: string, compatible with OCI image name specs.
  Name of the image to unpack. Must be available in the storepath, suffixed with `.oci.tgz`.
- value/archives/#/reference: string, compatible with OCI image reference specs.
  Reference inside the OCI archive to render. Must exist inside the OCI archive.

## JSON schema

```

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
