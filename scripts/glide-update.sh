#!/usr/bin/env bash
#
# Update vendored dedendencies.
#
set -e

if ! [[ "$PWD" = "$GOPATH/src/github.com/coreos/torcx" ]]; then
  echo "must be run from \$GOPATH/src/github.com/coreos/torcx"
  exit 1
fi

if [ ! $(command -v glide) ]; then
	echo "glide: command not found"
	exit 1
fi

if [ ! $(command -v glide-vc) ]; then
	echo "glide-vc: command not found"
	exit 1
fi

glide update --strip-vendor
glide-vc --only-code --no-tests --no-test-imports --use-lock-file
