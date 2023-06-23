MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_DIR := $(dir $(MAKEFILE_PATH))

DOCKER_COMPOSE_FILE_PATH := -f deployments/kubehound/docker-compose.yaml
DOCKER_COMPOSE_ENV_FILE_PATH := deployments/kubehound/.env
DEV_ENV_FILE_PATH := test/setup/.config


# get the latest commit hash in the short form
COMMIT := $(shell git rev-parse --short HEAD)
DATE := $(shell git log -1 --format=%cd --date=format:"%Y%m%d")
BUILD_VERSION := $(COMMIT)-$(DATA)

BUILD_FLAGS := -ldflags="-X github.com/DataDog/KubeHound/pkg/config.BuildVersion=$(BUILD_VERSION)"

# Need to save the MAKEFILE_LIST variable before the including the env var files
HELP_MAKEFILE_LIST := $(MAKEFILE_LIST)

# Loading docker .env file if present
ifneq (,$(wildcard $(DOCKER_COMPOSE_ENV_FILE_PATH)))
	include $(DOCKER_COMPOSE_ENV_FILE_PATH)
    export
endif

# Loading docker .env file if present
ifneq (,$(wildcard $(DEV_ENV_FILE_PATH)))
	include $(DEV_ENV_FILE_PATH)
    export
endif

ifneq ($(MAKECMDGOALS),system-test)
	ifeq (${KUBEHOUND_ENV}, prod)
		DOCKER_COMPOSE_FILE_PATH += -f deployments/kubehound/docker-compose.prod.yaml
	else ifeq (${KUBEHOUND_ENV}, dev)
		DOCKER_COMPOSE_FILE_PATH += -f deployments/kubehound/docker-compose.dev.yaml
	endif
endif

# No API key is being set
ifneq (${DD_API_KEY},)
    DOCKER_COMPOSE_FILE_PATH += -f deployments/kubehound/docker-compose.datadog.yaml
endif

UNAME_S := $(shell uname -s)
DOCKER_CMD := docker
ifndef DOCKER_CMD
	ifeq ($(UNAME_S),Linux)
		# https://docs.github.com/en/actions/learn-github-actions/variables
		ifneq (${CI},true)
			DOCKER_CMD := sudo docker
		endif
	endif
else
	DOCKER_CMD := ${DOCKER_CMD}
endif

RACE_FLAG_SYSTEM_TEST := "-race"
ifeq (${CI},true)
	RACE_FLAG_SYSTEM_TEST := ""
endif

all: build

.PHONY: generate
generate: ## Generate code the application
	go generate $(BUILD_FLAGS) ./...

.PHONY: build
build: generate ## Build the application
	cd cmd && go build $(BUILD_FLAGS) -o ../bin/kubehound kubehound/*.go

.PHONY: backend-rm
backend-rm: ## Delete the kubehound stack
	$(DOCKER_CMD) compose $(DOCKER_COMPOSE_FILE_PATH) rm -fvs 

.PHONY: backend-up
backend-up: ## Spawn the kubehound stack
	$(DOCKER_CMD) compose $(DOCKER_COMPOSE_FILE_PATH) up --force-recreate --build -d

.PHONY: backend-reset
backend-reset: ## Spawn the testing stack
	$(DOCKER_CMD) compose $(DOCKER_COMPOSE_FILE_PATH) rm -fvs 
	$(DOCKER_CMD) compose $(DOCKER_COMPOSE_FILE_PATH) up --force-recreate --build -d

.PHONY: test
test: ## Run the full suite of unit tests 
	cd pkg && go test -race $(BUILD_FLAGS) ./...

.PHONY: system-test
system-test: | backend-reset ## Run the system tests
	cd test/system && export KUBECONFIG=$(ROOT_DIR)/test/setup/${KIND_KUBECONFIG} && go test -v -timeout "60s" -count=1 -race ./...

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

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(HELP_MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'