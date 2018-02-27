# Release process

This document shows how to perform a torcx release and which tools are required for that.

Requirements:
 * git
 * make
 * gpg
 * docker

Environment:
 * New release version: `${NEWVER}` (e.g. `v2.1.0`)

Steps:
 1. Ensure you have a local clean checkout of current master branch:
    * `git checkout -f master`
    * `git reset --hard`
    * `git pull`
 1. Ensure master can be properly built and tested:
    * `make clean && make`
    * `make test`
    * `make ftest`
    * `make clean`
 1. Apply a signed tag to top commit and push it:
    * `git tag -s ${NEWVER} -m "torcx ${NEWVER}"`
    * `git push --tags`
 1. Build container image and push it:
    * `make clean && make container-amd64`
    * This will print the container image name. Double check it is in the form `quay.io/coreos/torcx:vx.y.z` and does NOT contain any `-dirty` or commit suffix.
    * `docker push quay.io/coreos/torcx:${NEWVER}`
 1. Perform a release on github:
    * Go to <https://github.com/coreos/torcx/releases> and add a new release.
    * Write a short summary of PRs and notable changes.
    * Publish the release.
