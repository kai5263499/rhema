FROM ubuntu:20.04

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ARG GO_VERSION=1.17.10
ARG PROTOC_VERSION=21.0

ENV CGO_ENABLED=1 CGO_CPPFLAGS="-I/usr/include"
ENV GOPATH=/go
ENV DEBIAN_FRONTEND=noninteractive
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH


COPY . /go/src/github.com/kai5263499/rhema

WORKDIR /go/src/github.com/kai5263499/rhema

RUN echo "Install apt packages" && \
	apt-get update && \
	apt-get install -y git gcc make curl unzip jq lame espeak-ng espeak-ng-data ffmpeg sox libsox-fmt-mp3 python ca-certificates

RUN	echo "Install youtube-dl" && \
	curl -sL https://yt-dl.org/downloads/latest/youtube-dl -o /usr/local/bin/youtube-dl && \
	chmod +x /usr/local/bin/youtube-dl && \
	/usr/local/bin/youtube-dl -U

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
