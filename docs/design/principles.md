# Design principles

These are the goals and ideas on which `torcx` is built upon:

1. early-boot integration: it is expected to run early in the boot phase, preparing other components in the system
1. systemd integration: it is expected to be used in conjunction with systemd units
1. profile-based: it is expected to handle pre-made profiles provided by user/vendor (as opposed to single packages)
1. low leve
1. local trust-boundary: authenticity/origin-trust is assumed for assets available on a local path
1. hard-pinning: it should be responsible of applying a completely-pinned configuration profile
1. fail hard, fail fast: it should either globally succeed or fail
1. (primarily) tailored to CL: it should solve the specific immutable-OS problem of decoupling services from base images
1. (primarily) tailored to docker: it should solve the specific problem of consumers pinning to specific docker versions

These are the non-goals and past-mistakes that `torcx` wants to avoid:

1. full packaging system: it should not be involved into solving versioned constraints
 
