# Well-known paths

Paths reserved for `torcx` usage.

Hardcoded:
* FusePath: `/run/metadata/torcx`
* VendorDir: `/usr/share/torcx/`

Configurable:
* `$TORCX_BASEDIR`: `/var/lib/torcx/`
* `$TORCX_RUNDIR`: `/var/run/torcx/`
* `$TORCX_CONFDIR`: `/etc/torcx/`

Derived from configurables (shown with defaults):
* BinDir: RunDir + `bin/` (`/var/run/torcx/bin/`)
* UnpackDir: RunDir + `unpack/` (`/var/run/torcx/unpack/`)
* Profile: ConfDir + `profile` (`/etc/torcx/profile`)
* AuthDir: 
  * (vendor) SharedDir + `auth.d/` (`/usr/share/torcx/auth.d/`)
  * (user) ConfDir + `auth.d/` (`/etc/torcx/auth.d/`)
* ProfileDir:
  * (vendor) SharedDir + `profile.d/` (`/usr/share/torcx/profile.d/`)
  * (user) ConfDir + `profile.d/` (`/etc/torcx/profile.d/`)

# Paths from environmental flags

## apply

* `$TORCX_STOREPATH`: additional store paths where to look for OCI archives (absolute, colon-separated)
* `$TORCX_SKIP`: if set to `true`, short-circuit `apply` into exiting early without applying any profile

# Fuse content

* `TORCX_PROFILE`: current running profile (default `vendor`)
* `TORCX_BINDIR`: current overlay with binaries, for `$PATH` usage (default `/var/run/torcx/bin/`)
