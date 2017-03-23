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
* CurrentProfile: RunDir + `profile` (`/var/run/torcx/profile`)
* NextProfile: ConfDir + `profile` (`/etc/torcx/profile`)
* StoreDir:
  * (vendor) VendorDir + `store/` (`/usr/share/torcx/store/`)
  * (user) BaseDir + `store/` (`/var/lib/torcx/store/`)
  * (runtime) `$TORCX_STOREPATH`
* AuthDir: 
  * (vendor) VendorDir + `auth.d/` (`/usr/share/torcx/auth.d/`)
  * (user) ConfDir + `auth.d/` (`/etc/torcx/auth.d/`)
* ProfileDir:
  * (vendor) VendorDir + `profiles.d/` (`/usr/share/torcx/profiles.d/`)
  * (user) ConfDir + `profiles.d/` (`/etc/torcx/profiles.d/`)

# Paths from environmental flags

## apply

* `$TORCX_STOREPATH`: additional store paths where to look for OCI archives (absolute, colon-separated)

# Fuse content

* `TORCX_PROFILE_NAME`: name of current running profile (default `vendor`)
* `TORCX_PROFILE_PATH`: path of current running profile (default `/var/run/torcx/profile`)
* `TORCX_BINDIR`: current overlay with binaries, for `$PATH` usage (default `/var/run/torcx/bin/`)
* `TORCX_UNPACKDIR`: current root of the unpacked tree (default `/var/run/torcx/unpack/`)
