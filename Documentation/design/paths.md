# Well-known paths

Paths reserved for `torcx` usage.

Hardcoded:
* SealFile: `/run/metadata/torcx`
* VendorDir: `/usr/share/torcx/`

Configurable via environmental flags:
* `$TORCX_BASEDIR`: `/var/lib/torcx/`
* `$TORCX_RUNDIR`: `/var/run/torcx/`
* `$TORCX_CONFDIR`: `/etc/torcx/`

Derived from configurables (shown with defaults):
* BinDir: RunDir + `bin/` (`/var/run/torcx/bin/`)
* UnpackDir: RunDir + `unpack/` (`/var/run/torcx/unpack/`)
* RunProfile: RunDir + `profile.json` (`/var/run/torcx/profile.json`)
* NextProfile: ConfDir + `next-profile` (`/etc/torcx/next-profile`)
* StoreDir:
  * (vendor) VendorDir + `store/` (`/usr/share/torcx/store/`)
  * (user) BaseDir + `store/` (`/var/lib/torcx/store/`)
  * (runtime) `$TORCX_STOREPATH`
* AuthDir: 
  * (vendor) VendorDir + `auth.d/` (`/usr/share/torcx/auth.d/`)
  * (user) ConfDir + `auth.d/` (`/etc/torcx/auth.d/`)
* ProfileDir:
  * (vendor) VendorDir + `profiles/` (`/usr/share/torcx/profiles/`)
  * (user) ConfDir + `profiles/` (`/etc/torcx/profiles/`)

# Paths from environmental flags

## apply

* `$TORCX_STOREPATH`: additional store paths where to look for addon images (ordered list of absolute paths, colon-separated)

# Seal file content

* `TORCX_PROFILE_NAME`: name of current running profile (default `vendor`)
* `TORCX_PROFILE_PATH`: path of current running profile (default `/var/run/torcx/profile.json`)
* `TORCX_BINDIR`: current overlay with binaries, for `$PATH` usage (default `/var/run/torcx/bin/`)
* `TORCX_UNPACKDIR`: current root of the unpacked tree (default `/var/run/torcx/unpack/`)
