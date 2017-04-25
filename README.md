<img align="left" width="70px" src="Documentation/torcx.png" />

# torcx - a boot-time addon manager

[![Apache](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)


torcx (pronounced _"torks"_) is a boot-time manager for system-wide ephemeral customization of Linux systems.
It has been built specifically to work with an immutable OS such as [Container Linux][coreos-cl] by CoreOS.

[coreos-cl]: https://coreos.com/releases/

torcx focuses on:
* providing a way for users to add additional binaries and services, even if not shipped in the base image
* allowing users to pin specific software versions, in a seamless and system-wide way
* supplying human- and machine-friendly interfaces to work with images and profiles

# <img src="https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Achtung.svg/2000px-Achtung.svg.png" alt="WARNING" width="25" height="25"> Disclaimer <img src="https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Achtung.svg/2000px-Achtung.svg.png" alt="NOTICE" width="25" height="25">

Torcx is currently in an experimental state. The API and CLI have no guarantees of stability, and the design is not yet finalized. Running torcx in production is not recommended.


## Getting started

This project provides a very lightweight add-ons manager for otherwise immutable distributions.
It applies collections of addon packages (named, respectively, "profiles" and "images") at boot-time, extracting them on the side of the base OS.

Profiles are simple JSON files, usually stored under `/etc/torcx/profiles/`, containing a set of image-references:

```json
{
  "kind": "profile-manifest-v0",
  "value": {
    "images": [
      {
        "name": "foo-binary",
        "reference": "0.1"
      }
    ]
  }
}

```

Image archives are looked up in several search paths, called "stores":
 1. Vendor store: usually on a read-only partition, it contains addons distributed together with the OS image
 1. User store: usually on a writable partition, it contains images provided by the user
 1. Runtime store: additional search path specified at runtime

At boot-time, torcx unpacks and propagates the addons defined in the active profile, specified in `/etc/torcx/next-profile`.
Once done, torcx seals the system into its new state and records its own metadata under `/run/metadata/torcx`.

## Example

Here is a short demo of torcx applying a profile with a single `socat` addon on top of a fresh Container Linux stable image.

[![asciicast](https://asciinema.org/a/115034.png)](https://asciinema.org/a/115034)

## License

torcx is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
