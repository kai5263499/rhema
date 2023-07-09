FROM ubuntu:22.04

ARG GO_VERSION=1.19.2
ARG PROTOC_VERSION=21.0

ENV CGO_ENABLED=1 CGO_CPPFLAGS="-I/usr/include"
ENV GOPATH=/go
ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /

COPY . /go/src/github.com/kai5263499/rhema

RUN echo "Install apt packages" && \
	apt-get update && \
	apt-get install -y git gcc make curl unzip jq lame espeak-ng espeak-ng-data ffmpeg sox libsox-fmt-mp3 python3 ca-certificates

RUN echo "Symlinking python to python3" && \
	ln -s $(which python3) /usr/bin/python

RUN	echo "Install yt-dlp" && \
	curl -sL https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
	chmod +x /usr/local/bin/yt-dlp && \
	/usr/local/bin/yt-dlp -U

RUN echo "Install golang" && \
	curl -sLO https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz && \
	tar -xf go${GO_VERSION}.linux-amd64.tar.gz && \
	rm -rf go${GO_VERSION}.linux-amd64.tar.gz

RUN cd /go/src/github.com/kai5263499/rhema && \
	mkdir -p out && \
	go build -o out/rhema cmd/rhema/main.go && \
    ldd out/apiserver | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname out/deps%); cp % out/deps%;' && \
	mv ./out/deps /deps && \
	mv ./out/rhema /rhema && \
	cd / && \
	rm -rf /go

ENV LOG_LEVEL="debug"
ENV PORT="8080"
ENV LOCAL_PATH="/data"
ENV TMP_PATH="/tmp"
ENV CHOWN_TO="1000"
ENV MIN_TEXT_BLOCK_SIZE="100"
ENV WORDS_PER_MINUTE="350"
ENV ESPEAK_VOICE="f5"
ENV ATEMPO="2.0"
ENV TITLE_LENGTH_LIMIT="40"

EXPOSE 8080

VOLUME [ "/tmp", "/data" ]

ENTRYPOINT [ "/rhema" ]