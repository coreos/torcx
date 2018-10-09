# Remote Manifest - v0

The configuration for a remote, locally stored on a node.
This is stored in a file called `remote.json` and located within a directory with the same name as the Torcx remote name.

## Schema
- kind (string, required)
- value (object, required)
  - base\_url (string, required)
  - keys (array, required, fixed-type, not-nil) - (object)
    - armored\_keyring (string)

## Entries

- `kind`: hardcoded to `remote-manifest-v0` for this schema revision. The type+version of this JSON manifest.
- `value`: object containing a single typed key-value. Manifest content.
- `value/base_url`: template with base URL for the remote. Supported protocols: "http", "https", "file".
- `value/keys/#`: array of single-type objects, arbitrary length. It contains trusted keys for signature verification.
- `value/keys/#/armored_keyring`: path to an ASCII-armored OpenPGP keyring, relative to the directory containing this remote manifest.

NOTE: `file://` URLs should generally only be used by offline remotes distributed as part of `/usr`, and controlled by the OS vendor.

## Templating

URL template in `base_url` is evaluated at runtime for simple variable substitution. Interpolated variables are:
 * `${COREOS_BOARD}`: board type (e.g. "amd64-usr")
 * `${COREOS_USR}`: path to a USR mountpoint (e.g. "/usr")
 * `${VERSION_ID}`: OS version (e.g. "1680.2.0")
 * `${ID}`: OS vendor ID (e.g. "coreos")

NOTE: `${COREOS_USR}` variable should generally only be used by offline remotes distributed as part of `/usr`, and controlled by the OS vendor.

## JSON schema

```json

{
  "$schema": "http://json-schema.org/draft-05/schema#",
  "type": "object",
  "properties": {
    "kind": {
      "type": "string",
      "enum": ["remote-manifest-v0"]
    },
    "value": {
      "type": "object",
      "properties": {
        "base_url": {
          "type": "string"
        },
        "keys": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "armored_keyring": {
                "type": "string"
              }
            }
          }
        }
      },
      "required": [
        "base_url",
        "keys"
      ]
    }
  },
  "required": [
    "kind",
    "value"
  ]
}

```
