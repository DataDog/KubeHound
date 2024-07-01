MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_DIR := $(dir $(MAKEFILE_PATH))

DEV_ENV_FILE_PATH := test/setup/.config
DEFAULT_KUBEHOUND_ENV := dev
SYSTEM_TEST_CMD := system-test system-test-clean

# get the latest commit hash in the short form
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")

BUILD_VERSION ?= $(shell git describe --match 'v[0-9]*' --dirty --always --tags)
BUILD_ARCH := $(shell go env GOARCH)
BUILD_OS := $(shell go env GOOS)

BUILD_FLAGS := -ldflags="-X github.com/DataDog/KubeHound/pkg/config.BuildVersion=$(BUILD_VERSION) -X github.com/DataDog/KubeHound/pkg/config.BuildArch=$(BUILD_ARCH) -X github.com/DataDog/KubeHound/pkg/config.BuildOs=$(BUILD_OS) -s -w"

# Need to save the MAKEFILE_LIST variable before the including the env var files
HELP_MAKEFILE_LIST := $(MAKEFILE_LIST)

# Loading docker .env file if present
ifneq (,$(wildcard $(DEV_ENV_FILE_PATH)))
	include $(DEV_ENV_FILE_PATH)
	export
endif

# Set default values if none of the above have set anything
ifndef KUBEHOUND_ENV
	KUBEHOUND_ENV := ${DEFAULT_KUBEHOUND_ENV}
endif

RACE_FLAG_SYSTEM_TEST := "-race"
ifeq (${CI},true)
	RACE_FLAG_SYSTEM_TEST := ""
endif

ifeq ($(OS),Windows_NT)
    DETECTED_OS = Windows
    DRIVE_PREFIX=C:
else
    DETECTED_OS = $(shell uname -s)
endif

ifeq ($(DETECTED_OS),Windows)
	BINARY_EXT=.exe
endif

# By default, all artifacts go to subdirectories under ./bin/ in the repo root
DESTDIR ?=

BUILDX_CMD ?= docker buildx

all: build

.PHONY: generate
generate: ## Generate code for the application
	go generate $(BUILD_FLAGS) ./...

.PHONY: build
build: ## Build the application
	go build $(BUILD_FLAGS) -o "$(or $(DESTDIR),./bin/build)/kubehound$(BINARY_EXT)" ./cmd/kubehound/

.PHONY: binary
binary:
	$(BUILDX_CMD) bake binary-with-coverage

.PHONY: lint
lint:
	$(BUILDX_CMD) bake lint

.PHONY: cross
cross: ## Compile the CLI for linux, darwin and windows (not working on M1)
	$(BUILDX_CMD) bake binary-cross

.PHONY: cache-clear
cache-clear: ## Clear the builder cache
	$(BUILDX_CMD) prune --force --filter type=exec.cachemount --filter=unused-for=24h

.PHONY: kubehound
kubehound: | build ## Prepare kubehound (build go binary, deploy backend)
	./bin/kubehound

.PHONY: test
test: ## Run the full suite of unit tests 
	cd pkg && go test -count=1 -race $(BUILD_FLAGS) ./...

.PHONY: system-test
system-test: | build ## Run the system tests
	./bin/build/kubehound dev system-tests
	cd test/system && export KUBECONFIG=$(ROOT_DIR)/test/setup/${KIND_KUBECONFIG} && go test -v -timeout "120s" -count=1 -race ./...

.PHONY: system-test-fast
system-test-fast: ## Run the system tests WITHOUT recreating the backend
	cd test/system && export KUBECONFIG=$(ROOT_DIR)/test/setup/${KIND_KUBECONFIG} && go test -v -timeout "60s" -count=1 -race ./...

.PHONY: system-test-clean
system-test-clean: | build ## Tear down the kubehound stack for the system-test
	./bin/build/kubehound dev system-tests --down

.PHONY: local-cluster-deploy
local-cluster-deploy: ## Create a kind cluster with some vulnerables resources (pods, roles, ...)
	bash test/setup/manage-cluster.sh destroy
	bash test/setup/manage-cluster.sh create
	bash test/setup/manage-cluster-resources.sh deploy

.PHONY: local-cluster-resource-deploy
local-cluster-resource-deploy: ## Deploy the attacks resources into the kind cluster
	bash test/setup/manage-cluster-resources.sh deploy

.PHONY: local-cluster-destroy
local-cluster-destroy: ## Destroy the local kind cluster
	bash test/setup/manage-cluster.sh destroy

.PHONY: sample-graph
sample-graph: | local-cluster-deploy build ## Create the kind cluster, start the backend, run the application, delete the cluster
	cd test/system && export KUBECONFIG=$(ROOT_DIR)/test/setup/${KIND_KUBECONFIG} && $(ROOT_DIR)/bin/build/kubehound
	bash test/setup/manage-cluster.sh destroy

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(HELP_MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: thirdparty-licenses
thirdparty-licenses: ## Generate the list of 3rd party dependencies and write to LICENSE-3rdparty.csv
	go get github.com/google/go-licenses
	go install github.com/google/go-licenses
	$(GOPATH)/bin/go-licenses csv github.com/DataDog/KubeHound/cmd/kubehound | sort > $(ROOT_DIR)/LICENSE-3rdparty.csv.raw
	python scripts/enrich-third-party-licences.py $(ROOT_DIR)/LICENSE-3rdparty.csv.raw > $(ROOT_DIR)/LICENSE-3rdparty.csv
	rm -f LICENSE-3rdparty.csv.raw

.PHONY: local-wiki
local-wiki: ## Generate and serve the mkdocs wiki on localhost
	poetry install || pip install mkdocs-material mkdocs-awesome-pages-plugin markdown-captions
	poetry run mkdocs serve || mkdocs serve

.PHONY: local-release
local-release: ## Generate release packages locally via goreleaser
	goreleaser release --snapshot --clean --config .goreleaser.yaml
