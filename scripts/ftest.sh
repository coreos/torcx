#!/usr/bin/env bash

# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

RKT_VER="1.26.0"
RKT_URL="https://github.com/rkt/rkt/releases/download/v${RKT_VER}/rkt-v${RKT_VER}.tar.gz"
RKT_STAGE1="coreos.com/rkt/stage1-coreos:${RKT_VER}"

if [ -n "${BUILDTAGS}" ]; then
    BUILDTAGS="-tags ${BUILDTAGS}"
fi

if [ ! -x "$(which rkt)" ]; then
    sudo mkdir -p /usr/local/bin/
    curl -L ${RKT_URL} |
    sudo tar -C /usr/local/bin --strip-components=1 -xazv rkt-v${RKT_VER}/rkt &&
    hash -r &&
    sudo rkt fetch --trust-keys-from-https=true ${RKT_STAGE1}
fi

echo "Running functional-tests:"
sudo -E env "GOPATH=$GOPATH" "PATH=$PATH" go test ${BUILDTAGS} -test.v ./ftests/
echo
