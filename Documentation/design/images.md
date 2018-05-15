# Image technical details

A torcx "[profile manifest][schemas]" specifies a set of addons (images and references) to be applied on a system.
Starting from a profile, torcx operates on a store of image archives.

## Archives

An archive is a squashfs filesystem containing a partial rootfs for a specific binary addon. For backwards compatibility, it may also be a gzipped tarball, but squashfs filesystems should be preferred.
Such archives are typically custom-built and tailored for torcx.

A torcx squashfs archive *MUST* be a [version 4.0](https://github.com/torvalds/linux/blob/v4.16/Documentation/filesystems/squashfs.txt) squashfs filesystem archive. It *MUST* be compressed using either gzip or lz4.

## References

Image references may entail special values reserved by vendors, such as `com.coreos.cl`.

# Dynamic binaries and libraries

Binaries and libraries from images are usually not extracted into global system paths, to keep torcx side-effects contained to specific directories.
As such, dynamic artifacts shipped into images should typically take additional steps to be fully portable according to torcx logic.

The following caveats should be addressed, usually at build-time, to ensure full interoperability:
* system-wide libraries: if a binary requires system libraries, the build system must ensure compatibility with the target runtime environment.
* library load path: it is recommended to use relative rpath on binaries, anchored at `$ORIGIN`. E.g. a `/usr/bin/hello` binary may need to have a `$ORIGIN/../lib/` origin to find custom libs.
* binary paths: binaries are exposed under a custom `$PATH` entry owned by torcx, and should be looked up that way. If an absolute path is required, it can be constructed using [runtime metadata][paths]

Static binaries, should not be affected by any specific caveats.

# System-integration paths

`torcx-generator` consumes the "[image manifest][schemas]" embedded in each image to identify artifacts that need to be integrated with the system.
It currently looks for:
* binaries
* systemd service units
* systemd network units
* sysusers files
* tmpfiles files
* udev rules

[schemas]: ./schemas.md
[paths]: ./paths.md
