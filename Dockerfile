FROM ubuntu:18.04

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

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
	curl -sL https://yt-dl.org/latest -o /usr/local/bin/youtube-dl && \
	chmod +x /usr/local/bin/youtube-dl && \
	/usr/local/bin/youtube-dl -U

RUN echo "Install golang" && \
	curl -sLO https://dl.google.com/go/go1.13.3.linux-amd64.tar.gz && \
	tar -xf go1.13.3.linux-amd64.tar.gz && \
	mv go /usr/local && \
	rm -rf go1.13.3.linux-amd64.tar.gz

RUN echo "Caching golang modules" && \
	go mod download && \
	go get -u github.com/swaggo/swag/cmd/swag

RUN	echo "Install protoc tools" && \
	go get -u github.com/golang/protobuf/protoc-gen-go && \
	curl -sLO https://github.com/google/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip && \
    unzip protoc-3.7.1-linux-x86_64.zip -d protoc3 && \
    mv protoc3/bin/* /usr/local/bin/ && \
    mv protoc3/include/* /usr/local/include/ && \
    rm -rf protoc3 protoc-3.7.1-linux-x86_64.zip
