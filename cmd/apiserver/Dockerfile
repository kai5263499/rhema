FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/apiserver && \
    go build && \
    ldd apiserver | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM scratch

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV LOG_LEVEL="info"
ENV PORT="8080"
ENV MQTT_BROKER=""
ENV MQTT_CLIENT_ID="apiserver"
ENV REDIS_HOST=""
ENV REDIS_PORT="6379"
ENV REDIS_GRAPH_KEY="rhema-content"
ENV GOOGLE_APPLICATION_CREDENTIALS=""
ENV AUTH0_CLIENT_ID=""
ENV AUTH0_DOMAIN=""
ENV AUTH0_CLIENT_SECRET=""
ENV AUTH0_CALLBACK_URL="http://localhost:3000/callback"
ENV ENABLE_GRAPHIQL=false

EXPOSE 8080

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/apiserver/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/apiserver/apiserver /apiserver

ENTRYPOINT [ "/apiserver" ]