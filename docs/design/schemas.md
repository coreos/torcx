# Schemas and types

# Profile selector

* `/etc/torcx/profile`: single plaintext string, `EOF`-terminated
* Profile: non-empty string, allowed char in regexp `^[a-zA-Z._-]{1,512}$`

# Runtime metadata

* `/run/metadata/torcx`: "key=value" environment variables, each line `\n`-terminated

# Profile manifest

* Schema at `schemas/profile-manifest-v<n>.json` where highest `<n>` is the current version.
