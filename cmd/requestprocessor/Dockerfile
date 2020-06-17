FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/requestprocessor && \
    go build && \
    ldd requestprocessor | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM ubuntu:18.04

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV LOG_LEVEL="info"
ENV MQTT_BROKER=""
ENV MQTT_CLIENT_ID="requestprocessor"
ENV MIN_TEXT_BLOCK_SIZE="100"
ENV TMP_PATH="/tmp"
ENV WORDS_PER_MINUTE="350"
ENV ESPEAK_VOICE="f5"
ENV ATEMPO="2.0"
ENV TITLE_LENGTH_LIMIT="40"

RUN echo "Install apt packages" && \
	apt-get update && \
    apt-get install -y lame espeak-ng espeak-ng-data ffmpeg python ca-certificates

COPY --from=builder /usr/local/bin/youtube-dl /usr/local/bin/youtube-dl

RUN /usr/local/bin/youtube-dl -U

COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/requestprocessor/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/requestprocessor/requestprocessor /requestprocessor

ENTRYPOINT [ "/requestprocessor" ]