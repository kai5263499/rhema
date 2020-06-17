FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/apiserver && \
    go build && \
    ldd apiserver | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM scratch

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV LOG_LEVEL="info"
ENV MQTT_BROKER=""

COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/apiserver/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/apiserver/apiserver /apiserver

ENTRYPOINT [ "/apiserver" ]