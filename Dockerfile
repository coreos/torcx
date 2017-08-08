FROM scratch

LABEL org.label-schema.vendor="CoreOS Inc. <coreos-dev@googlegroups.com>" \
      org.label-schema.vcs-url="https://github.com/coreos/torcx"

COPY bin/amd64/torcx /usr/sbin/torcx
