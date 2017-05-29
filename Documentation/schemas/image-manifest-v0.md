# Image Manifest - v0

An "image manifest" is a JSON data structure consumed by torcx and usually provided inside an image as a file under `.torcx/manifest.json`.
It contains multiple lists of assets (organized by type) to be propagated on the host system.

## Schema

- kind (string, required)
- value (object, required)
  - bin (array of strings, optional)
  - units (array of strings, optional)

Note: The list of optional assets types will likely grow in the future. This is a non-breaking change, and does not require bumping the `kind` field.

## Entries

- kind: hardcoded to `image-manifest-v0` for this schema revision.
  The type+version of this JSON manifest.
- value: object containing a single typed key-value.
  Manifest content.
- value/bin: array of string, arbitrary length.
  List of absolute paths for binaries to be propagated under torcx bin directory.
- value/network: array of string, arbitrary length.
  List of absolute paths of networkd units to be propagated under networkd runtime directory. This can reference single unit-files as well as directories (e.g. for ".conf" dropins)
- value/units: array of string, arbitrary length.
  List of absolute paths of units to be propagated under systemd runtime directory. This can reference single unit-files as well as directories (e.g. for ".wants" and ".requires")

## JSON schema

```json
{
  "$schema": "http://json-schema.org/draft-05/schema#",
  "type": "object",
  "properties": {
    "kind": {
      "type": "string",
      "enum": ["image-manifest-v0"]
    },
    "value": {
      "type": "object",
      "properties": {
        "bin": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "network": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "units": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
		"sysusers": {
          "type": "array",
          "items": {
            "type": "string"
          }
        },
        "tmpfiles": {
          "type": "array",
          "items": {
            "type": "string"
          }
        }

      }
    }
  },
  "required": [
    "kind",
    "value"
  ]
}
```

## Example

See [examples/image-manifest.json] for a sample.
