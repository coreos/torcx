# Design principles

These are the goals and ideas on which `torcx` is built upon:

1. early-boot integration: it is expected to run early in the boot phase, preparing other components in the system
1. systemd integration: it is expected to be used in conjunction with systemd units
1. profile-based: it is expected to handle pre-made profiles provided by user/vendor (as opposed to single packages)
1. low level, machine-friendly: it should provide the equivalent of a package system which can be easily driven by external containerized agents
1. local trust-boundary: authenticity/origin-trust is assumed for assets available on a local path
1. hard-pinning: it should be responsible of applying a completely-pinned configuration profile
1. fail hard, fail fast: it should either globally succeed or fail
1. (primarily) tailored to CL: it should solve the specific immutable-OS problem of decoupling services from base images
1. (primarily) tailored to docker: it should solve the specific problem of consumers pinning to specific docker versions

These are the non-goals and out-of-scope topics that `torcx` wants to avoid:

1. versioned dependency resolution: a profile is simple collection of images to be applied, without any explicit relationship
1. full packaging system: it should not be involved into complex upgrade paths
1. upgrading and downgrading packages at runtime: profile activation is an atomic operation performed at-most-once at boot-time
1. pre-removal and post-installation custom logic: changes performed by torcx are (mainly) volatile and are meant to only last for a single boot
1. defining a custom package format: torcx just handles squashfs rootfs archives

## User Stories

A few motivating examples, presented as user stories.

### 1: Docker version
A user wants to select a docker version by ignition at first boot time.

#### Design:
In the ignition config, you will have to write a profile manifest and select that
profile. The vendor packages are available in /usr, and custom packages are written
to disk already (network share, installed by Ignition, etc...).

### 2: Docker version, redux
The Kubernetes Version Operator wants to select a docker version at upgrade
time on an already running machine.

#### Design:
Torcx binary is available in a container with the necessary host paths bind-mounted in. The KVO executes a sequence of torcx commands to select the desired Docker version.

### 3: Fleet
A user wants to install fleet at first boot time, but wants automatic docker upgrades.

#### Design:
Similar to story 1, but with fleet as an additional entry in the archives list,
with a reserved reference (matchin the one used in OS-vendored packages).

### 4: Up-to-date
The user wants whatever Container Linux ships, running the latest recommended version of Docker.

#### Design:
The user specifies 'vendor' profile in Ignition, or nothing at all.

### 5: Barebone OS
The user doesn't want torcx to run at all, and doesn't want it to potentially prevent machines from booting.

#### Design:
User masks/disables the unit.
