FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/contentbot && \
    go build && \
    ldd contentbot | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM scratch

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV LOG_LEVEL="info"
ENV MQTT_BROKER=""
ENV MQTT_CLIENT_ID="contentbot"
ENV SLACK_TOKEN=""

COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/contentbot/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/contentbot/contentbot /contentbot

ENTRYPOINT [ "/contentbot" ]