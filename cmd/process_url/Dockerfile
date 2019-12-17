FROM kai5263499/rhema-builder as builder

COPY / /go/src/github.com/kai5263499/rhema

RUN cd /go/src/github.com/kai5263499/rhema/cmd/process_url && \
    go build && \
    ldd process_url | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname deps%); cp % deps%;'

FROM scratch

LABEL MAINTAINER="Wes Widner <kai5263499@gmail.com>"

ENV AWS_DEFAULT_REGION=""
ENV MIN_TEXT_BLOCK_SIZE="100"
ENV S3_BUCKET=""
ENV LOCAL_PATH="/tmp"
ENV WORDS_PER_MINUTE="350"
ENV ESPEAK_VOICE="f5"

COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/process_url/deps /
COPY --from=builder /go/src/github.com/kai5263499/rhema/cmd/process_url/process_url /process_url

ENTRYPOINT [ "/process_url" ]