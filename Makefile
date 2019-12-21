# Build the rhema-builder image with go and other associated libraries
builder-image:
	docker build -t kai5263499/rhema-builder .

# Build the rhema-process-url image for processing content from the command line
process-urls-image:
	docker build -t kai5263499/rhema-process-url -f cmd/process_url/Dockerfile .

# Build the rhema-scrape image for processing content from the command line
scrape-image:
	docker build -t kai5263499/rhema-scrape -f cmd/scrape/Dockerfile .

# Generate go stubs from proto definitions. This should be run inside of an interactive container
go-protos:
	protoc -I proto/ proto/*.proto --go_out=plugins=grpc:generated

# Run an interactive shell for development and testing
exec-interactive:
	docker run -it --rm \
	-e PULSE_SERVER=172.17.0.1 \
	-e S3_BUCKET="${S3_BUCKET}" \
	-e AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION}" \
	-e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
	-e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
	-e SLACK_TOKEN="${SLACK_TOKEN}" \
	-e LOG_LEVEL="${LOG_LEVEL}" \
	-v ${LOCAL_DEV_PATH}:/go/src/github.com/kai5263499 \
	-v ${LOCAL_CONTENT_PATH}:/data \
	--tmpfs /tmp:exec \
	-w /go/src/github.com/kai5263499/rhema/cmd/contentbot \
	kai5263499/rhema-builder bash

test:
	go test

.PHONY: image exec-interactive protos test
