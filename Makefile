
DOCKER_REPO ?= kai5263499

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

MODULES=$(subst cmd/,,$(wildcard cmd/*))

GIT_COMMIT := $(shell git rev-parse HEAD | cut -c 1-8)
GIT_BRANCH := $(shell git branch --show-current)

which = $(shell which $1 2> /dev/null || echo $1)

GO_PATH := $(call which,go)
$(GO_PATH):
	$(error Missing go)

MOQ_PATH := $(call which,moq)
$(MOQ_PATH):
	@$(GO_PATH) install github.com/matryer/moq@latest

DOCKER_PATH := $(call which,docker)
$(DOCKER_PATH):
	$(error Missing docker)

OAPI_PATH := $(call which,oapi-codegen)
$(OAPI_PATH):
	$(error Missing oapi-codegen: https://github.com/deepmap/oapi-codegen)

# linux/arm64,

image:
	@$(DOCKER_PATH) buildx build --load --platform linux/amd64 -t ${DOCKER_REPO}/rhema -f .docker/Dockerfile .

TOOLS=$(wildcard tools/*)
.PHONY: ${TOOLS}
${TOOLS}:
	@echo "running $(subst tools/,,$@)"
	@$(GO_PATH) run ./$@

# Options to pass to docker compose. Like an .env var file to use
DOCKER_COMPOSE_OPTIONS:=""

# shutdown & cleanup
container/environment/clean: down
	@rm -f $(ROOT_DIR)/active_profiles

# Generate go stubs from proto definitions. This should be run inside of an interactive container
PROTOC_PATH := $(call which,protoc)
$(PROTOC_PATH):
	$(error Missing protoc: https://developers.google.com/protocol-buffers/docs/gotutorial)
protos:
	@$(PROTOC_PATH) -I proto/ proto/*.proto --go_out=generated

PUML_PATH := $(call which,plantuml)
$(PUML_PATH):
	$(error Missing plantuml: https://plantuml.com/starting)
puml:
	@$(PUML_PATH) -t png -o . $(wildcard doc/*.puml)

LINTER_PATH := $(call which,golangci-lint)
$(LINTER_PATH):
	$(error Missing golangci: https://golangci-lint.run/usage/install)
lint:
	@rm -rf ./vendor
	@$(GO_PATH) mod vendor
	export GOMODCACHE=./vendor
	@$(LINTER_PATH) run

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

api/v1: $(OAPI_PATH)
	@$(GO_PATH) generate -v ./internal/v1/...
.PHONY: api/v1

unit/tests: test

TEST_ARGS := -cover -v -count=1 -timeout 15m
TEST_FILTER?=.

test:
	@$(GO_PATH) test $(TEST_ARGS) $(sort $(dir $(shell find . -type f -name '*.go' | grep -ivw 'cmd\|domain\|tools\|vendor'))) -tags=!intg
test/%:
	@$(GO_PATH) test $(TEST_ARGS) ./... -run $(@F)
intg/tests:
	@$(GO_PATH) test $(TEST_ARGS) $(sort $(dir $(shell find . -type f -name '*.go' | grep -ivw 'cmd\|domain\|tools\|vendor'))) -tags=intg -args online
test/cmd/%:
	@$(GO_PATH) test $(TEST_ARGS) -tags=$(@F) ./cmd/$(@F)/... -run ${TEST_FILTER}

GO_JUNIT_REPORT_PATH := $(call which,go-junit-report)
$(GO_JUNIT_REPORT_PATH):
	$(error Missing go-junit-report)

test/junit:
	@make test | $(GO_JUNIT_REPORT_PATH) -set-exit-code > testOutput.xml

mock: $(MOQ_PATH)
	@$(GO_PATH) generate ./...
.PHONY: mock

clean: down
	@rm -rf ./vendor
	@rm -rf out
	@find . -name "*.gen.go" -exec rm -rf {} \;
	@rm -f $(ROOT_DIR)/active_profiles

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

OPTIONS?=""

which = $(shell which $1 2> /dev/null || echo $1)
