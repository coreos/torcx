FROM debian:stretch-slim

MAINTAINER CoreOS Inc. <coreos-dev@googlegroups.com>

ADD bin/amd64/torcx /usr/sbin/torcx

RUN  mkdir -p /run/metadata /var/lib/torcx /etc/torcx && \
     mkdir -p /etc/containers && echo '{"default": [{"type": "insecureAcceptAnything"}]}' > /etc/containers/policy.json


USER root:root
CMD /bin/bash
