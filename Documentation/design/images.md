# Image technical details

torcx profiles specify images and references.
torcx operates on a store of image archives.

An archive is a gzip-compressed tar file, containing a partial rootfs for a specific binary addon.
torcx images are shipped as tgz archives, however they are typically custom built and tailored for torcx usages.

Image references may contain custom values as specified by the provider.
It is recommend to use semver-compatible tags for normal release and vendor-scoped names for special purpose.

For example, a special reference may be used by vendors to specify the latest current version, such as `com.coreos.cl`.

# Dynamic binaries and libraries

Binaries and libraries from images are usually not extracted into global system paths, to keep torcx side-effects contained to specific directories.
As such, dynamic artifacts shipped into images should typically take additional steps to be fully portable according to torcx logic.

The following caveats should be addressed, usually at build-time, to ensure full interoperability:
* library load path: it is recommended to use relative rpath on binaries, anchored at `$ORIGIN`. E.g. a `/usr/bin/hello` binary may need to have a `$ORIGIN/../lib/` origin to find custom libs.
* binary paths: binaries are exposed under a custom `$PATH` entry owned by torcx, and should be looked up that way. If an absolute path is required, it can be constructed using [runtime metadata][torcx-paths]

Static binaries, should not be affected by any specific caveats.

[torcx-paths]: https://github.com/coreos/torcx/blob/master/Documentation/design/paths.md

# System-integration paths

torcx will scan images to identify artifacts that need to be integrated with the system.
It currently looks for:
* binaries: files in `/bin:/usr/bin:/sbin:/usr/sbin` are symlinked into `$TORCX_BINDIR` and exposed system-wide that way
* systemd units: files in `/usr/lib/systemd/system` are extracted/symlinked into systemd-specific path as transient units (typically `/run/systemd/system/`)

TODO(lucab): sysctl - tmpfile - sysusers units, the latter ones have some GC concerns to be explored.
