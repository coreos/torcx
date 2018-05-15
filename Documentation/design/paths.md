# Well-known paths

Paths reserved for `torcx` usage.

Hardcoded:
* SealFile: `/run/metadata/torcx`
* VendorDir: `/usr/share/torcx/`
* OemDir: `/usr/share/oem/torcx/`

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
  * (versioned-oem) OemDir + `store/` + CurOSVer (`/usr/share/oem/torcx/store/<CurOSVer>/`)
  * (oem) OemDir + `store/` (`/usr/share/oem/torcx/store`)
  * (versioned-user) BaseDir + `store/` + CurOSVer (`/var/lib/torcx/store/<CurOSVer>/`)
  * (user) BaseDir + `store/` (`/var/lib/torcx/store/`)
  * (runtime) `$TORCX_STOREPATH`
* ProfileDir:
  * (vendor) VendorDir + `profiles/` (`/usr/share/torcx/profiles/`)
  * (oem) OemDir + `profiles/` (`/usr/share/oem/torcx/profiles/`)
  * (user) ConfDir + `profiles/` (`/etc/torcx/profiles/`)
* RemotesDir:
  * (vendor) VendorDir + `remotes/` (`/usr/share/torcx/remotes/`)
  * (oem) OemDir + `remotes/` (`/usr/share/oem/torcx/remotes/`)
  * (user) ConfDir + `remotes/` (`/etc/torcx/remotes/`)

# Paths from environmental flags

## apply

* `$TORCX_STOREPATH`: additional store paths where to look for addon images (ordered list of absolute paths, colon-separated)

# Seal file content

* `TORCX_LOWER_PROFILES`: array of names of lower vendor/oem profiles, separated by `:` (default `vendor:oem`)
* `TORCX_UPPER_PROFILE`: name of current running user profile (default ``)
* `TORCX_PROFILE_PATH`: path of current running profile (default `/run/torcx/profile.json`)
* `TORCX_BINDIR`: current overlay with binaries, for `$PATH` usage (default `/run/torcx/bin/`)
* `TORCX_UNPACKDIR`: current root of the unpacked tree (default `/run/torcx/unpack/`)
