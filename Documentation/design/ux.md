# User Interface and Experience

## Concepts

 * *torcx*: a CLI tool organized in subcommand
 * *archive*: an torcx.tgz archive containing addon binaries and assets
 * *profile*: a manifest specifying a set of images to be applied on top of the base OS
 * *vendor profile*: the default profile, shipped hardcoded inside the OS
 * *custom profile*: a thrid-party profile *not* shipped together with the OS
 * *vendor store*: a set of archives shipped by the OS, possibly residing in a RO area
 * *custom store*: a set of archives provided by third-parties, possibly residing in RW area
 

## Subcommands

### Profile commands

```
torcx profile new [--from=<FNAME> | --from-next] --name=<PNAME>|--file=<PATH>
```

Creates a new profile PNAME or file PATH. 

If PNAME is specified, it is created in `$TORCX_CONFDIR/profiles.d/NAME`. It 
must not already exist as a profile. 

If `--from` is specified, the new profile is a duplicate of profile FNAME.

If `--from-next` is specified, the profile is a duplicate of whichever profile is
marked as active for the next boot.

IF no `--from*` argument is specified, the created profile is empty.

```
torcx profile set-next <NAME>
```

Switches to profile NAME on next boot.

```
torcx profile list
```

Lists the available profiles, indicating the currently-booted and profile selected
for next boot.

```
torcx profile use-image [--allow=missing] --name=<PNAME>|--file=<PATH> <NAME>:<REFERENCE>
```

Adds image refered by NAME and reference REFERENCE to the given profile called
PNAME or at path PATH.

If the image does not exist, this will abort unless --allow=missing is supplied.

```
torcx profile check [--name=<PNAME> | --file=<PATH>]
```

Check that the profile named by PNAME or file PATH is apply-able - that all images
exist in the stores. An apply-able profile will have an exit code of 0.

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

