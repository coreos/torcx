# Boot-time torcx integration

torcx is an addon manager which is executed at a very early stage in the Linux boot process.
It works as a systemd generator, which takes care of unpacking and propagating assets contained inside addon images.

## Multicall binary

torcx is a multicall binary which is aware of its invocation context (ie. binary name) and automatically switches semantics based on that.
It currently knows about the following names:
 * `torcx`: this is the main run-time entrypoint. It provides further subcommands and flags, and is meant to be invoked by users.
 * `torcx-generator`: this is the main boot-time entrypoint for the generator. It does not provide any subcommands or flags.

## Installation paths

torcx needs to be available locally on the system to perform its task. These are the typical installation paths:
 * the generator component needs to be named `torcx-generator` and available under one of systemd lookup paths. This typically means `/usr/lib/systemd/system-generators/torcx-generator` or `/etc/systemd/system-generators/torcx-generator`.
 * the runtime component needs to be named `torcx` and can be located at any executable location, possibly somewhere inside `$PATH` like `/usr/bin/torcx`.

## Generator configuration

`torcx-generator` does not accept any subcommands, command-line flags, or environmental options due to systemd generator protocol.
However its behavior can be optionally tweaked at runtime via a configuration file, located a `/etc/torcx/config.json`.
The configuration path can be change by providing a `torcx_config=` parameter to kernel command-line, pointing it to a different file.
