TORCX_FILES := $(shell find rootfs)

.PHONY: all symlinks
all: addon_01.torcx.tgz addon_02.torcx.squashfs symlinks

addon_01.torcx.tgz: $(TORCX_FILES)
	tar -C rootfs -czf $@ .

addon_02.torcx.squashfs: $(TORCX_FILES)
	mksquashfs rootfs $@

# Since 'make' can't handle a target with a ':' in it, but torcx addons are
# defined to contain colons, create non-colon targets and then just have
# symlinks floating around for them.
symlinks:
	ln -sf "addon_01.torcx.tgz" "addon:01.torcx.tgz"
	ln -sf "addon_02.torcx.squashfs" "addon:02.torcx.squashfs"


clean:
	rm -f addon_01.torcx.tgz addon_02.torcx.squashfs
