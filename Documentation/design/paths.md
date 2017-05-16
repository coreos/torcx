# Well-known paths

Paths reserved for `torcx` usage.

Hardcoded:
* SealFile: `/run/metadata/torcx`
* VendorDir: `/usr/share/torcx/`

Configurable via environmental flags:
* `$TORCX_BASEDIR`: `/var/lib/torcx/`
* `$TORCX_RUNDIR`: `/run/torcx/`
* `$TORCX_CONFDIR`: `/etc/torcx/`

Derived from configurables (shown with defaults):
* BinDir: RunDir + `bin/` (`/run/torcx/bin/`)
* UnpackDir: RunDir + `unpack/` (`/run/torcx/unpack/`)
* RunProfile: RunDir + `profile.json` (`/run/torcx/profile.json`)
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

* `TORCX_VENDOR_PROFILE`: name of the base vendor profile (default `vendor`)
* `TORCX_USER_PROFILE`: name of current running user profile (default ``)
* `TORCX_PROFILE_PATH`: path of current running profile (default `/run/torcx/profile.json`)
* `TORCX_BINDIR`: current overlay with binaries, for `$PATH` usage (default `/run/torcx/bin/`)
* `TORCX_UNPACKDIR`: current root of the unpacked tree (default `/run/torcx/unpack/`)
