# User Interface and Experience

## Concepts

 * *torcx*: a CLI tool organized in subcommand
 * *archive*: an OCI image-layout archive in tar-gz format
 * *profile*: a manifest specifying a set of archives to be applied on top of the base OS
 * *vendor profile*: the default profile, shipped hardcoded inside the OS
 * *custom profile*: a thrid-party profile *not* shipped together with the OS
 * *vendor store*: a set of archives shipped by the OS, possibly residing in a RO area
 * *custom store*: a set of archives provided by third-parties, possibly residing in RW area
 

## Subcommands

### Profile commands
```
torcx profile new [--from=<FNAME>] <NAME>
```

Creates a new custom staging profile NAME. It must not already exist as a profile. If 
`--from` is specified, the new profile is a duplicate of profile FNAME.

The profile is created in `$TORCX_CONFDIR/profile.d/NAME`.

```
torcx profile rm <NAME>
```

Deletes profile NAME. The specified profile must not be the active one, and must
be user-created.

```
torcx profile select <NAME>
```

Switches to profile NAME on next boot.

```
torcx profile list
```

Lists the available profiles, indicating the currently-booted and profile selected
for next boot.

```
torcx profile use <PNAME> NAME[:TAG|@DIGEST]
```

Adds package refered by NAME (and other possible references) to profile PNAME.

OPEN QUESTION: can `use` act on the current profile?

```
torcx profile check <PNAME>
```

Check that the profile named by PNAME is apply-able - that all packages and references
exist in the stores. Report any packages that are missing.

### Package commands
```
torcx package fetch NAME[:TAG|@DIGEST]
```

`fetch` fetches a package into the user store.

```
torcx package cp <PATH>
```

`package cp` copies a package at a given path in to the user store. If <PATH>
is `-`, then the package contents are received over stdin.

```
torcx package rm NAME
```

`package rm` removes a package from the user store.

```
torcx package list [NAME]
```

List all packages and the available references.

If NAME is specified, only list the references for that package.

```
torcx package gc
```

`gc` cleans up unreferenced OCI archives from user store.

### Other commands
```
torcx apply
```

`apply` applies current profile to the machine.


This is meant to be used exactly once-per-boot, and blows a fuse after successful setup.

