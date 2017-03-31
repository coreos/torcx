# torcx - system-wide manager for bundles and profiles

torcx (pronounced _"torks"_) is a system-wide manager for bundles and profiles, specific to CoreOS ContainerLinux (CCL).

torcx's goals include:
* providing a way for distribution maintainers and users to add additional binaries and services to CCL, without shipping them in the base images
* allowing users to run specific software versions, in a seamless and system-wide way
* supply a CLI tool for manipulating bundles and runtime profiles

## Overview

This project provides a very lightweight add-ons manager for otherwise immutable distributions, such as CCL.
It handles collection of packages (here called respectively "profiles" and "bundles") at boot-time, overlaying them on top of the base OS image.

As such, torcx fulfill two main roles:
* at boot-time, it activates a specific profile by unpacking all requested bundles and making them available system-wide
* at runtime, it provides a human- and machine-friendly interface to work with bundles and profiles

Contrary to traditional packaging systems and add-on managers, torcx scope is quite limited and explicitly does not support:
* upgrading and downgrading packages at runtime. Profile activation is a single atomic operation performed once, at boot-time
* pre-removal and post-installation custom logic. Changes performed by torcx are volatile and are meant to only last for a single boot
* defining a custom package format. torcx just handles OCI image-layout archives
* versioned dependency resolution. A profile is simple collection of bundles to be applied, without any implicit relationship
