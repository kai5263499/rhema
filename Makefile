all-images:
	docker build -t kai5263499/rhema-builder .
	docker build -t kai5263499/rhema-process-url -f cmd/processurl/Dockerfile .
	docker build -t kai5263499/rhema-scrape -f cmd/scrape/Dockerfile .
	docker build -t kai5263499/rhema-bot -f cmd/contentbot/Dockerfile .
	docker build -t kai5263499/rhema-processor -f cmd/processor/Dockerfile .
	docker build -t kai5263499/rhema-storage -f cmd/storage/Dockerfile .
	docker build -t kai5263499/rhema-apiserver -f cmd/apiserver/Dockerfile .

# Generate go stubs from proto definitions. This should be run inside of an interactive container
go-protos:
	protoc -I proto/ proto/*.proto --go_out=generated

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
	cd cmd/processor && go build
	cd cmd/storage && go build

test:
	go test

.PHONY: exec-interactive go-protos test all-services all-images
