
DOCKER_REPO := kai5263499

MODULES=$(subst cmd/,,$(wildcard cmd/*))
TEST_FILTER?=.

GIT_COMMIT := $(shell git rev-parse HEAD | cut -c 1-8)
GIT_BRANCH := $(shell git branch --show-current)

which = $(shell which $1 2> /dev/null || echo $1)

GO_PATH := $(call which,go)
$(GO_PATH):
	$(error Missing go)

MOQ_PATH := $(call which,moq)
$(MOQ_PATH):
	@$(GO_PATH) install github.com/matryer/moq@latest

PUML_PATH := $(call which,plantuml)
$(PUML_PATH):
	$(error Missing plantuml: https://plantuml.com/starting)

OAPI := $(call which,oapi-codegen)

FILTER_OUT = $(foreach v,$(2),$(if $(findstring $(1),$(v)),,$(v)))

IMAGE_MODULES=$(call FILTER_OUT,image/processor, $(addprefix image/,$(MODULES)))
.PHONY: ${IMAGE_MODULES}

OUT_MODULES=$(addprefix out/,$(MODULES))
.PHONY: ${OUT_MODULES}

image/builder:
	docker build -t ${DOCKER_REPO}/rhema-builder -f Dockerfile.builder .

image/all: image/builder
	for image in ${IMAGE_MODULES}; do \
        make $$image ; \
    done

${IMAGE_MODULES}:
	docker rmi -f ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT}
	docker build -t ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} -f Dockerfile.cmd --build-arg service_name=$(subst image/,,$@) .
	docker tag ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} ${DOCKER_REPO}/$(subst image/,,$@):${GIT_BRANCH}
	docker rmi $$(docker images -f "dangling=true" -q) || true

image/processor:
	docker rmi -f ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT}
	docker build -t ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} -f Dockerfile.cmd.processor .
	docker tag ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} ${DOCKER_REPO}/$(subst image/,,$@):${GIT_BRANCH}
	docker rmi $$(docker images -f "dangling=true" -q) || true

${OUT_MODULES}:
	@$(GO_PATH) build -gcflags=all="-N -l" -o $@ ./cmd/$(subst out/,,$@)

out/all:
	for out in ${OUT_MODULES}; do \
        make $$out ; \
    done

TOOLS=$(wildcard tools/*)
.PHONY: ${TOOLS}

${TOOLS}:
	@echo "running $(subst tools/,,$@)"
	@$(GO_PATH) run ./$@

# Generate go stubs from proto definitions. This should be run inside of an interactive container
protos:
	protoc -I proto/ proto/*.proto --go_out=generated

# Run an interactive shell for development and testing
container/interactive:
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
	-e GOOGLE_APPLICATION_CREDENTIALS=/tmp/gcp/service-account-file.json \
	-e AUTH0_CLIENT_ID=${AUTH0_CLIENT_ID} \
	-e AUTH0_CLIENT_SECRET=${AUTH0_CLIENT_SECRET} \
	-e AUTH0_DOMAIN=${AUTH0_DOMAIN} \
	-e AUTH0_CALLBACK_URL=${AUTH0_CALLBACK_URL} \
	--tmpfs /tmp:exec \
	-p 8090:8080 \
	-w /go/src/github.com/kai5263499/rhema/cmd/apiserver \
	kai5263499/rhema-builder bash

mktmp:
	mkdir -p /tmp/mqtt-data /tmp/mqtt-log /tmp/redisgraph

docker/mosquitto/up: mktmp
	docker run -d \
	--name mosquitto \
	-p 1883:1883 -p 9001:9001 \
	-v /tmp/mqtt-data:/mosquitto/data \
	-v /tmp/mqtt-log:/mosquitto/log \
	-v mosquitto.conf:/mosquitto/config/mosquitto.conf \
	eclipse-mosquitto

docker/mosquitto/down:
	docker rm -f mosquitto

docker/redisgraph/up: mktmp
	docker run -d \
	--name regisgraph \
	-p 6397:6397 \
	-v /tmp/regisgraphdata:/data
	redislabs/redisgraph

docker/redisgraph/down:
	docker rm -f redisgraph

test/%:
	go test -v -count=1 ./... -run $(@F)

intg/tests:
	go test -cover -count=1 $(sort $(dir $(shell find . -type f -name '*.go' | grep -ivw 'cmd\|domain\|tools\|vendor'))) -tags=intg -args online

unit/tests: test

test:
	go test -cover -count=1 $(sort $(dir $(shell find . -type f -name '*.go' | grep -ivw 'cmd\|domain\|tools\|vendor'))) -tags=!intg

test/cmd/%:
	go test -v -timeout 15m -count=1 -tags=$(@F) ./cmd/$(@F)/... -run ${TEST_FILTER}

ds/install:
	helm install rhema ./helm 

ds/uninstall:
	helm delete rhema

puml:
	plantuml -t png -o . $(wildcard doc/*.puml)

LINTER_PATH := $(call which,golangci-lint)
$(LINTER_PATH):
	$(error Missing golangci: https://golangci-lint.run/usage/install)
lint:
	@rm -rf ./vendor
	@$(GO_PATH) mod vendor
	export GOMODCACHE=./vendor
	golangci-lint run

# MUST pin oapi-codegen to match go.mod
OAPI_PATH := $(call which,oapi-codegen)
$(OAPI_PATH):
	$(error Missing oapi-codegen: go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@v1.9.0)

internal/%/server.gen.go: api/%.yaml 
	@$(OAPI_PATH) -package $(notdir $(@D)) -include-tags="$*" -generate spec,server $< > $@

internal/%/types.gen.go: api/%.yaml 
	@$(OAPI_PATH) -package $(notdir $(@D)) -generate types $< > $@

internal/%/origin.gen.go: api/%.yaml 
	@$(OAPI_PATH) -package $(notdir $(@D)) -include-tags="origin" -generate client $< > $@

%/types.gen.go: api/%.yaml 
	@$(OAPI_PATH) -package agentconfig -include-tags="$*" -generate types $< > $@

%/client.gen.go: api/%.yaml 
	@$(OAPI_PATH) -package agentconfig -include-tags="$*" -generate client $< > $@

api/v1: $(OAPI_PATH) \
	v1/client.gen.go \
	v1/types.gen.go \
	internal/v1/origin.gen.go \
	internal/v1/server.gen.go \
	internal/v1/types.gen.go 
.PHONY: api/v1

GO_JUNIT_REPORT_PATH := $(call which,go-junit-report)
$(GO_JUNIT_REPORT_PATH):
	$(error Missing go-junit-report)

TEST_ARGS := -cover -v -count=1
test: 
	@$(GO_PATH) test $(TEST_ARGS) ./internal/... -tags=!intg

test/junit:
	@make test | $(GO_JUNIT_REPORT_PATH) -set-exit-code > testOutput.xml

MOQ_PATH := $(HOME)/go/bin/moq
$(MOQ_PATH):
	$(GO_PATH) install github.com/matryer/moq@latest

mock: $(MOQ_PATH)
	@$(GO_PATH) generate ./...
.PHONY: mock

clean:
	@rm -rf ./vendor
	@rm -rf out
	@find . -name "*.gen.go" -exec rm -rf {} \;

