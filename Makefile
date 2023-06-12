
DOCKER_REPO := kai5263499

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

FILTER_OUT = $(foreach v,$(2),$(if $(findstring $(1),$(v)),,$(v)))

IMAGE_MODULES=$(call FILTER_OUT,image/processor, $(addprefix image/,$(MODULES)))
.PHONY: ${IMAGE_MODULES}

OUT_MODULES=$(addprefix out/,$(MODULES))
.PHONY: ${OUT_MODULES}

image/builder:
	@$(DOCKER_PATH) build -t ${DOCKER_REPO}/rhema-builder -f .docker/Dockerfile.builder .

image/all: image/builder
	for image in ${IMAGE_MODULES}; do \
        make $$image ; \
    done

${IMAGE_MODULES}:
	@$(DOCKER_PATH) rmi -f ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT}
	@$(DOCKER_PATH) build -t ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} -f .docker/Dockerfile.cmd --build-arg service_name=$(subst image/,,$@) .
	@$(DOCKER_PATH) tag ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} ${DOCKER_REPO}/$(subst image/,,$@):${GIT_BRANCH}
	@$(DOCKER_PATH) rmi $$(docker images -f "dangling=true" -q) || true

image/processor:
	@$(DOCKER_PATH) rmi -f ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT}
	@$(DOCKER_PATH) build -t ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} -f Dockerfile.cmd.processor .
	@$(DOCKER_PATH) tag ${DOCKER_REPO}/$(subst image/,,$@):${GIT_COMMIT} ${DOCKER_REPO}/$(subst image/,,$@):${GIT_BRANCH}
	@$(DOCKER_PATH) rmi $$(docker images -f "dangling=true" -q) || true

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

# Run an interactive shell for development and testing
container/interactive:
	@$(DOCKER_PATH) run -it --rm \
	-e BUCKET="${BUCKET}" \
	-e LOG_LEVEL="${LOG_LEVEL}" \
	-e CHANNELS="${CHANNELS}" \
	-v ${DEV_PATH}:/go/src/github.com/kai5263499 \
	-v ${LOCAL_PATH}:/data \
	--tmpfs /tmp:exec \
	-p 8090:8080 \
	-w /go/src/github.com/kai5263499/rhema/cmd/apiserver \
	kai5263499/rhema-builder bash

HELM_PATH := $(call which,helm)
$(HELM_PATH):
	$(error Missing helm)
ds/install:
	@$(HELM_PATH) install rhema ./helm 

ds/uninstall:
	@$(HELM_PATH) delete rhema

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

define COMPOSE_DEFS
-f $(ROOT_DIR)/.docker/docker-compose-redis.yml \
-f $(ROOT_DIR)/.docker/docker-compose-statsd.yml \
-f $(ROOT_DIR)/.docker/docker-compose-kafka.yml \
-f $(ROOT_DIR)/.docker/docker-compose-apiserver.yml \
-f $(ROOT_DIR)/.docker/docker-compose-traefik.yml
endef

# startup
.PHONY: up
up:
	@cd $(ROOT_DIR) && \
	COMPOSE_PROFILES="$$(cat active_profiles | tr '\n' ',' | sed 's/\(.*\),/\1/')" \
	docker compose \
	  $(COMPOSE_DEFS) \
	  $$(echo "$(OPTIONS)") \
	  up -d

# shutdown
.PHONY: down
down:
	@cd $(ROOT_DIR) && \
	COMPOSE_PROFILES="$(ALL_PROFILES)" \
	docker compose \
	  $(COMPOSE_DEFS) \
	  $$(echo "$(OPTIONS)") \
	  down

.PHONY: view
view:
	@echo "active profiles:"
	@cat $(ROOT_DIR)/active_profiles

.PHONY: reset
reset: down up

del-%:
	@make del/$(subst del-,,$(@F))

del/%:
	@sed -i '/^$(subst del/,,$(@F))$$/d' $(ROOT_DIR)/active_profiles

add-%:
	@make add/$(subst add-,,$(@F))

add/%:
	@echo "$(subst add/,,$(@F))" >> $(ROOT_DIR)/active_profiles

certs/setup:
	wget https://github.com/FiloSottile/mkcert/releases/download/v1.4.4/mkcert-v1.4.4-linux-amd64
	sudo mv mkcert-v1.4.4-linux-amd64 /usr/local/bin/mkcert && chmod +x /usr/local/bin/mkcert
	mkcert -install
	mkcert -cert-file certs/local-cert.pem -key-file certs/local-key.pem "docker.localhost" "*.docker.localhost" "domain.local" "*.domain.local"
