FROM kai5263499/rhema-builder as builder

ARG service_name

COPY / /go/src/github.com/kai5263499/rhema

RUN make out/${service_name} && \
    ldd out/${service_name} | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname out/deps%); cp % out/deps%;'

FROM scratch

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV LOG_LEVEL="info"
ENV PORT="8080"
ENV MQTT_BROKER=""
ENV MQTT_CLIENT_ID="${service_name}"
ENV REDIS_HOST=""
ENV REDIS_PORT="6379"
ENV REDIS_GRAPH_KEY="rhema-content"
ENV AUTH0_CLIENT_ID=""
ENV AUTH0_DOMAIN=""
ENV AUTH0_CLIENT_SECRET=""
ENV AUTH0_CALLBACK_URL="http://localhost:3000/callback"
ENV BUCKET=""
ENV LOCAL_PATH="/data"
ENV TMP_PATH="/tmp"
ENV CHOWN_TO="1000"
ENV GOOGLE_APPLICATION_CREDENTIALS=""
ENV MIN_TEXT_BLOCK_SIZE="100"

EXPOSE 8080

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/kai5263499/rhema/out/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/out/${service_name} /${service_name}

ENTRYPOINT [ "/${service_name}" ]