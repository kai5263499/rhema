FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/apiserver && \
    go build && \
    ldd apiserver | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM ubuntu:18.04

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV AWS_DEFAULT_REGION=""
ENV MIN_TEXT_BLOCK_SIZE="100"
ENV S3_BUCKET=""
ENV LOCAL_PATH="/data"
ENV TMP_PATH="/tmp"
ENV WORDS_PER_MINUTE="350"
ENV ESPEAK_VOICE="f5"
ENV ATEMPO="2.0"
ENV CHOWN_TO="1000"
ENV LOG_LEVEL="info"

RUN echo "Install apt packages" && \
	apt-get update && \
    apt-get install -y lame espeak-ng espeak-ng-data ffmpeg python ca-certificates

COPY --from=builder /usr/local/bin/youtube-dl /usr/local/bin/youtube-dl
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/apiserver/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/apiserver/apiserver /apiserver

ENV MIN_TEXT_BLOCK_SIZE="100"
ENV BUCKET=""
ENV LOCAL_PATH="/data"
ENV TMP_PATH="/tmp"
ENV WORDS_PER_MINUTE="350"
ENV ESPEAK_VOICE="f5"
ENV ATEMPO="2.0"
ENV CHOWN_TO="1000"
ENV LOG_LEVEL="info"
ENV SLACK_TOKEN=""
ENV ELASTICSEARCH_URL=""
ENV TITLE_LENGTH_LIMIT="40"
ENV GOOGLE_APPLICATION_CREDENTIALS=""

ENTRYPOINT [ "/apiserver" ]