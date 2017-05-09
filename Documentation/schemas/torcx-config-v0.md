# torcx Config - v0

torcx common config is a JSON data structure that can be optionally provided by an external party (e.g. an user) to influence torcx behavior.
It is typically stored under `/etc/torcx/config.json`, but a custom path can be specified with a `torcx_config=` setting at kernel command-line.

## Schema

- kind (string, required)
- value (object, required)
  - base_dir (string, optional)
  - conf_dir (string, optional)
  - run_dir (string, optional)
  - store_paths (array of string, optional)

## Entries

- kind: hardcoded to `torcx-config-v0` for this schema revision.
  The type+version of this JSON manifest.
- value: object containing a single typed key-value.
  Config content.
- value/base_dir: optional string.
  Custom path to override base directory.
- value/conf_dir: optional string.
  Custom path to override config directory.
- value/run_dir: optional string.
  Custom path to override runtime directory.
- value/store_paths: optional array of strings.
  A list of store paths to add to the lookup paths.
