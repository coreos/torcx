# torcx - 

torcx (pronounced _"torks"_) is a system-wide manager for images and profiles, specific to CoreOS ContainerLinux (CCL).

torcx's goals include:
* providing a way for distribution maintainers and users to add additional binaries and services to CCL, without shipping them in the base images
* allowing users to run specific software versions, in a seamless and system-wide way
* supply a CLI tool for manipulating images and runtime profiles

## Overview

This project provides a very lightweight add-ons manager for otherwise immutable distributions, such as CCL.
It handles collection of packages (named, respectively, "profiles" and "images") at boot-time, overlaying them on top of the base OS image.

As such, torcx fulfill two main roles:
* at boot-time, it activates a specific profile by unpacking all requested images and making them available system-wide
* at runtime, it provides a human- and machine-friendly interface to work with images and profiles

Contrary to traditional packaging systems and add-on managers, torcx scope is quite limited and explicitly does not support:
* upgrading and downgrading packages at runtime. Profile activation is a single atomic operation performed once, at boot-time
* pre-removal and post-installation custom logic. Changes performed by torcx are volatile and are meant to only last for a single boot
* defining a custom package format. torcx just handles OCI image-layout archives
* versioned dependency resolution. A profile is simple collection of images to be applied, without any implicit relationship
