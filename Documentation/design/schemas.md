# Schemas and types

This document describes format of manifests and configuration files related to torcx.

# Profile selector

* `/etc/torcx/profile`: plaintext string (leading and trailing spaces are ignored)
* profile name: non-empty string, allowed characters in regexp `^[a-zA-Z._-]{1,512}$`

# Runtime metadata

* `/run/metadata/torcx`: "key=value" environment variables, each line `\n`-terminated

# JSON manifests

Configuration files and assets manifests are in JSON format.
Existing [manifest schemas][schemas] are:
* Image manifest (`schemas/image-manifest-v<n>.json`): describes the content of an image.
* Profile manifest (`schemas/profile-manifest-v<n>.json`): describes the set of images in a profile.
* Torcx config (`schemas/torcx-config-v<n>.json`): global torcx configuration.

[schemas]: ../schemas
