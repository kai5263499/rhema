FROM ubuntu:22.04

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ARG GO_VERSION=1.19.2
ARG PROTOC_VERSION=21.0

ENV CGO_ENABLED=1 CGO_CPPFLAGS="-I/usr/include"
ENV GOPATH=/go
ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

COPY . /go/src/github.com/kai5263499/rhema

WORKDIR /go/src/github.com/kai5263499/rhema

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
	mv go /usr/local && \
	rm -rf go${GO_VERSION}.linux-amd64.tar.gz

RUN echo "Cache golang modules" && \
	go mod download && \
	go mod vendor

RUN	echo "Install protoc" && \
	go get -u github.com/golang/protobuf/protoc-gen-go && \
	curl -sLO https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}-linux-x86_64.zip && \
    unzip protoc-${PROTOC_VERSION}-linux-x86_64.zip -d protoc3 && \
    mv protoc3/bin/* /usr/local/bin/ && \
    mv protoc3/include/* /usr/local/include/ && \
    rm -rf protoc3 protoc-${PROTOC_VERSION}-linux-x86_64.zip
