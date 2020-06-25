FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/storage && \
    go build && \
    ldd storage | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM scratch

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV LOG_LEVEL="info"
ENV MQTT_BROKER=""
ENV MQTT_CLIENT_ID="storage"
ENV BUCKET=""
ENV LOCAL_PATH="/data"
ENV TMP_PATH="/tmp"
ENV CHOWN_TO="1000"
ENV GOOGLE_APPLICATION_CREDENTIALS=""

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/storage/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/storage/storage /storage

ENTRYPOINT [ "/storage" ]