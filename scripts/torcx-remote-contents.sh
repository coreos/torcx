#!/usr/bin/env bash
# Copyright 2018 Red Hat.
# Licensed under the Apache License, Version 2.0 (the "License").

## Scan a directory of assets and print the corresponding
## `torcx-remote-contents-v1` manifest:
## $ torcx-remote-contents.sh -p . > torcx_remote_contents.json


set -eo pipefail

ASSETS_PATH=${ASSETS_PATH:-.}

# Index map: "name" -> [ list of hashes ]
declare -A addons
# Property maps: "name:hash" -> property"
declare -A formats
declare -A locations
declare -A versions

## Scan for images

scan_assets() {
  for path in $("${BIN_FIND}" "${ASSETS_PATH}" -type f \( -name '*:*.torcx.tgz' -o -name '*:*.torcx.squashfs' \) -printf '%P\n'); do
    local img namever name version format shahash namehash seen
    img="$(echo "${path}" | rev | cut -d'/' -f 1 | rev)"
    namever="$(echo "${img}" | rev | cut -d'.' -f 3- | rev)"

    # Extract image properties from filepath
    name="$(echo "${namever}" | cut -d':' -f 1)"
    version="$(echo "${namever}" | cut -d':' -f 2)"
    format="$(echo "${img}" | rev | cut -d'.' -f 1 | rev)"
    shahash="sha512-$("${BIN_SHASUM}" "${ASSETS_PATH}"/"${path}" | cut -d' ' -f 1)"

    # Record properties in keyed maps
    namehash="${name}:${shahash}"
    seen="${addons[${name}]}"
    addons["${name}"]="${shahash} ${seen}"
    formats["${namehash}"]="${format}"
    locations["${namehash}"]="${path}"
    versions["${namehash}"]="${version}"
  done
}

## Print manifest to stdout

print_manifest() {
  local imagecomma=""

  # Print fixed header
  "${BIN_PRINTF}" "${HEADER}"

  for name in "${!addons[@]}"; do
    local vercomma=""

    "${BIN_PRINTF}" "${imagecomma}"
    imagecomma=","
    # Interpolate and print image header
    "${BIN_PRINTF}" "${NAME_HEADER}" "${name}"

    for hash in ${addons[$name]}; do
      "${BIN_PRINTF}" "${vercomma}"
      vercomma=","
      local namehash="${name}:${hash}"

      # Interpolate and print image-version template
      "${BIN_PRINTF}" "${IMAGE_TEMPLATE}" \
       "${versions[${namehash}]}" \
       "${formats[${namehash}]}" \
       "${locations[${namehash}]}" \
       "${hash}"
    done

    # Print fixed image footer
    "${BIN_PRINTF}" "${NAME_FOOTER}"

  done

  # Print fixed manifest footer
  "${BIN_PRINTF}" "${FOOTER}"
}


## Templates

HEADER='{
  "kind": "torcx-remote-contents-v1",
  "value": {
    "images": ['

NAME_HEADER='
      {
        "name": "%s",
        "versions": ['

IMAGE_TEMPLATE='
          {
            "version": "%s",
            "format": "%s",
            "location": "%s",
            "hash": "%s"
          }'

NAME_FOOTER='
        ]
      }
'

FOOTER="    ]
  }
}
"

## Script body

while getopts ":p:" OPTION
do
    case $OPTION in
        p) ASSETS_PATH="${OPTARG}" ;;
        *) echo "usage: $0 [-p ASSETS_PATH]" >&2
           exit 1 ;;
    esac
done

if ! BIN_FIND=$(which find 2> /dev/null); then
  echo "no find binary found" >&2
  exit 1
fi
if ! BIN_PRINTF=$(which printf 2> /dev/null); then
  echo "no printf binary found" >&2
  exit 1
fi
if ! BIN_SHASUM=$(which sha512sum 2> /dev/null); then
  echo "no sha512sum binary found" >&2
  exit 1
fi

if [[ ! -d "${ASSETS_PATH}" ]]; then
  echo "assets directory not found" >&2
  exit 1
fi

scan_assets

print_manifest
