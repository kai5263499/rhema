FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/contentbot && \
    go build && \
    ldd contentbot | tr -s '[:blank:]' '\n' | grep '^/' | \
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
ENV SLACK_TOKEN=""

RUN echo "Install apt packages" && \
	apt-get update && \
    apt-get install -y lame espeak-ng espeak-ng-data ffmpeg python curl

RUN	echo "Install youtube-dl" && \
	curl -sL https://github.com/ytdl-org/youtube-dl/releases/download/2019.11.28/youtube-dl -o /usr/local/bin/youtube-dl && \
	chmod +x /usr/local/bin/youtube-dl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/contentbot/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/contentbot/contentbot /contentbot

ENTRYPOINT [ "/contentbot" ]