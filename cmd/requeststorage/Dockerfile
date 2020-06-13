FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/requeststorage && \
    go build && \
    ldd requeststorage | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM ubuntu:18.04

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV BUCKET=""
ENV LOCAL_PATH="/data"
ENV TMP_PATH="/tmp"
ENV CHOWN_TO="1000"
ENV LOG_LEVEL="info"
ENV GOOGLE_APPLICATION_CREDENTIALS=""

RUN echo "Install apt packages" && \
	apt-get update && \
    apt-get install -y lame espeak-ng espeak-ng-data ffmpeg python ca-certificates

COPY --from=builder /usr/local/bin/youtube-dl /usr/local/bin/youtube-dl
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/requeststorage/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/requeststorage/requeststorage /requeststorage

ENTRYPOINT [ "/requeststorage" ]