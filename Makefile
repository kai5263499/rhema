# Build the rhema-builder image with go and other associated libraries
builder-image:
	docker build -t kai5263499/rhema-builder .

# Build the rhema-process-url image for processing content from the command line
process-urls-image:
	docker build -t kai5263499/rhema-process-url -f cmd/process_url/Dockerfile .

# Build the rhema-scrape image for processing content from the command line
scrape-image:
	docker build -t kai5263499/rhema-scrape -f cmd/scrape/Dockerfile .

# Build the rhema-bot image for processing content from slack messages
contentbot-image:
	docker build -t kai5263499/rhema-bot -f cmd/contentbot/Dockerfile .

# Generate go stubs from proto definitions. This should be run inside of an interactive container
go-protos:
	protoc -I proto/ proto/*.proto --go_out=plugins=grpc:generated

# Run an interactive shell for development and testing
exec-interactive:
	docker run -it --rm \
	-e BUCKET="${BUCKET}" \
	-e MQTT_BROKER="${MQTT_BROKER}" \
	-e SLACK_TOKEN="${SLACK_TOKEN}" \
	-e LOG_LEVEL="${LOG_LEVEL}" \
	-e CHANNELS="${CHANNELS}" \
	-e GOOGLE_APPLICATION_CREDENTIALS="/tmp/gcp/service-account-file.json" \
	-v ${DEV_PATH}:/go/src/github.com/kai5263499 \
	-v ${LOCAL_PATH}:/data \
	-v ${GOOGLE_APPLICATION_CREDENTIALS}:/tmp/gcp/service-account-file.json \
	--tmpfs /tmp:exec \
	-p 8090:8080 \
	-w /go/src/github.com/kai5263499/rhema/cmd/apiserver \
	kai5263499/rhema-builder bash

exec-contentbot:
	docker run -it --rm \
	-e BUCKET="${BUCKET}" \
	-e SLACK_TOKEN="${SLACK_TOKEN}" \
	-e LOG_LEVEL="${LOG_LEVEL}" \
	-v ${LOCAL_CONTENT_PATH}:/data \
	-v ${LOCAL_TMP_PATH}:/tmp \
	kai5263499/rhema-bot

exec-apiserver:
	docker run -it --rm \
	-p 8081:8080 \
	-e BUCKET="${BUCKET}" \
	-e LOG_LEVEL="${LOG_LEVEL}" \
	-e GOOGLE_APPLICATION_CREDENTIALS="/tmp/gcp/service-account-file.json" \
	-e ELASTICSEARCH_URL=${ELASTICSEARCH_URL} \
	-v ${LOCAL_CONTENT_PATH}:/data \
	-v ${LOCAL_TMP_PATH}:/tmp \
	-v ${GOOGLE_APPLICATION_CREDENTIALS}:/tmp/gcp/service-account-file.json \
	kai5263499/rhema-apiserver

elasticsearch:
	docker run -d \
	-p 9200:9200 \
	-p 9300:9300 \
	--name elasticsearch \
	-e "discovery.type=single-node" \
	docker.elastic.co/elasticsearch/elasticsearch:7.6.1

mosquitto:
	docker run -d \
	--name mosquitto \
	-p 1883:1883 -p 9001:9001 \
	-v /tmp/mqtt-data:/mosquitto/data \
	-v /tmp/mqtt-log:/mosquitto/log \
	-v mosquitto.conf:/mosquitto/config/mosquitto.conf \
	eclipse-mosquitto

build-all:
	cd cmd/apiserver && go build
	cd cmd/contentbot && go build
	cd cmd/processurl && go build
	cd cmd/scrape && go build
	cd cmd/processurl  && go build
	cd cmd/requestprocessor && go build
	cd cmd/requeststorage && go build

test:
	go test

.PHONY: image exec-interactive protos test all-services
