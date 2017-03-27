* Are hashes, tags, and references interchangeable? How so?
	* where are they not?
	* What is the store layout on disk? Will it need to be frozen for Ignition?
	* What if there is a conflicting tag?
* How does one specify that a single package should diverge from vendor (including through updates)
	* motivating example: I want a pinned version of fleet, but auto-updating everything else
	* possible solutions:
		* add some kind of `from` directive to the profile
		* Use a reserved vendor reference name (tag, b/c mutable)
* How to handle systemd services?
	Imagine that you get the docker tarball, and there are monstrous pile of services inside.
	* Load them as a transient unit
	* Symlink to our own path
	* Symlink each one as a transient unit
