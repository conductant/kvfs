# KVFS

FROM alpine:3.3

MAINTAINER david.chung@docker.com

RUN apk --update add bash
RUN apk --update add fuse
RUN rm -rf /var/cache/apk/*

ADD build/linux-amd64/kvfs /usr/local/bin/kvfs

# Print out version so we know the build info.
RUN kvfs version

# Need to change the permission of the /dev/fuse file
RUN chmod a+rw /dev/fuse

ENTRYPOINT ["kvfs"]