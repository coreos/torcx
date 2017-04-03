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

The profile is created in `$TORCX_CONFDIR/profiles.d/NAME`.

```
torcx profile rm <NAME>
```

Deletes profile NAME. The specified profile must not be the one selected for next boot, and must
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
torcx profile use <PNAME> NAME(:TAG|@DIGEST)
```

Adds image refered by NAME and TAG or DIGEST to profile PNAME. One of TAG or
DIGEST are required.

```
torcx profile check <PNAME>
```

Check that the profile named by PNAME is apply-able - that all images
exist in the stores. Report any packages that are missing.

### Bundle commands

```
torcx image fetch NAME[:TAG|@DIGEST]
```

`fetch` fetches an image into the user store.

```
torcx image list [NAME]
```

List all images in the store.

If NAME is specified, only list the references for that image name.

```
torcx image list-unused
```

Lists the unused archives in the user store.

### Other commands

```
torcx apply
```

`apply` applies current profile to the machine.

This is meant to be used exactly once-per-boot, and blows a fuse after successful setup.

