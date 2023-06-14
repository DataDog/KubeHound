BUILD_VERSION=dev-snapshot

MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
ROOT_DIR := $(dir $(MAKEFILE_PATH))

DOCKER_COMPOSE_FILE_PATH := -f test/system/docker-compose.yaml -f test/system/docker-compose.local.yaml

# https://docs.github.com/en/actions/learn-github-actions/variables
ifeq (${CI},true)
    DOCKER_COMPOSE_FILE_PATH := -f test/system/docker-compose.yaml
endif

all: build

.PHONEY: build
build:
	cd cmd && go build -ldflags="-X pkg/config.BuildVersion=$(BUILD_VERSION)" -o ../bin/kubehound kubehound/*.go

.PHONY: infra-rm
infra-rm:
	docker compose -f $(DOCKER_COMPOSE_FILE_PATH) rm -fvs 

.PHONY: infra-up
infra-up:
	docker compose -f $(DOCKER_COMPOSE_FILE_PATH) up -d

.PHONY: test
test:
	$(MAKE) infra-rm
	$(MAKE) infra-up
	cd pkg && go test ./...

.PHONY: system-test
system-test: 
	$(MAKE) infra-rm
	$(MAKE) infra-up
	cd test/system && go test -v -timeout "60s" -race ./...

.PHONY: local-cluster-reset
local-cluster-reset:
	$(MAKE) local-cluster-destroy
	$(MAKE) local-cluster-setup
	$(MAKE) local-cluster-config-deploy

.PHONY: local-cluster-deploy
local-cluster-deploy:
	$(MAKE) local-cluster-setup
	$(MAKE) local-cluster-config-deploy

.PHONY: local-cluster-config-deploy
local-cluster-config-deploy:
	bash test/setup/manage-cluster-resources.sh deploy

.PHONY: local-cluster-config-delete
local-cluster-config-delete:
	bash test/setup/manage-cluster-resources.sh delete

.PHONY: local-cluster-create
local-cluster-create:
	bash test/setup/manage-cluster.sh create

.PHONY: local-cluster-destroy
local-cluster-destroy:
	bash test/setup/manage-cluster.sh destroy
