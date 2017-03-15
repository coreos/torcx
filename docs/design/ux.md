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

```
torcx new --base=[vendor|empty|<pbase>] --name=<pname>
```

`new` creates a new custom staging profile "pname". It must not already exist as a profile.

```
torcx fetch coreos/flannel@hash:deadbeef
```

`fetch` fetches a package into the user store.

```
torcx add/remove coreos/flannel@hash:deadbeef
```

A pair of commands to add/remove a local package to the staging area

```
torcx commit
```

Finalize current staging area, will become “user” on next boot

```
torcx select --type=[vendor|user] --name=<pname>
```

`select` switches to profile "pname" for next boot.

```
torcx gc
```

`gc` cleans up unreferenced OCI archives from user store.

```
torcx apply
```

`apply` applies current profile.
This is meant to be used exactly once-per-boot, and blows a fuse after successful setup.
