# Bundle technical details

torcx acts on a set of bundles, as specified by the current profile manifest.

A bundle is a gzip-compressed tar archive, containing an OCI image-archive layout as specified by [OCI image-spec][image-layout].

[image-layout]: https://github.com/opencontainers/image-spec/blob/master/image-layout.md

# Annotations and references

torcx bundles are shipped as OCI archives, however they are typically custom built and tailored for torcx usages.
As such, an OCI image must be annotated with a reserved annotation in order to be used by torcx, as such:
```
"com.coreos.torcx.bundle.v0": "yes"
```

Image tags (references) may contain custom value as specified by the bundle provider.
It is recommend to use semver-compatible tags for normal release and vendor-scoped names for special purpose.

For example, a special tag may be used by vendors to specify the latest current version, such as `com.coreos.cl`.

# Dynamic binaries and libraries

Binaries and libraries from bundles are usually not extracted into global system paths, to keep torcx side-effects contained to specific directories.
As such, dynamic artifacts shipped into bundles should typically take additional steps to be fully portable according to torcx logic.

The following caveats should be addressed, usually at build-time, to ensure full interoperability:
* library load path: it is recommended to use relative rpath on binaries, anchored at `$ORIGIN`. E.g. a `/usr/bin/hello` binary may need to have a `$ORIGIN/../lib/` origin to find custom libs.
* binary paths: binaries are exposed under a custom `$PATH` entry owned by torcx, and should be looked up that way. If an absolute path is required, it can be constructed using [runtime metadata][torcx-paths]

Static binaries, should not be affected by any specific caveats.

[torcx-paths]: https://github.com/coreos/torcx/blob/master/Documentation/design/paths.md

# System-integration paths

torcx will scan bundles to identify artifacts that need to be integrated with the system.
It currently looks for:
* binaries: files in `/bin:/usr/bin:/sbin:/usr/sbin` are symlinked into `$TORCX_BINDIR` and exposed system-wide that way
* systemd units: files in `/usr/lib/systemd/system` are extracted/symlinked into systemd-specific path as transient units (typically `/run/systemd/system/`)

TODO(lucab): sysctl - tmpfile - sysusers units, the latter ones have some GC concerns to be explored.
