#!/usr/bin/make -f
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app

# for dockerized protobuf tools
DOCKER := $(shell which docker)
BUF_IMAGE=bufbuild/buf@sha256:3cb1f8a4b48bd5ad8f09168f10f607ddc318af202f5c057d52a45216793d85e5 #v1.4.0
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(BUF_IMAGE)
HTTPS_GIT := https://github.com/CosmWasm/wasmd.git

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
empty = $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(empty),$(comma),$(build_tags),ledger)


# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=pellcore \
		  -X github.com/cosmos/cosmos-sdk/version.AppName=pellcored \
		  -X github.com/cosmos/cosmos-sdk/version.ClientName=pellclientd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X github.com/0xPellNetwork/aegis/pkg/constant.Name=pellcored \
	      -X github.com/0xPellNetwork/aegis/pkg/constant.Version=$(VERSION) \
		  -X github.com/CosmWasm/wasmd/app.Bech32Prefix=pell \
		  -X github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)' -trimpath
TEST_DIR?="./..."
TEST_BUILD_FLAGS := -tags goleveldb,ledger
HSM_BUILD_FLAGS := -tags goleveldb,ledger,hsm_test

export DOCKER_BUILDKIT := 1

clean: clean-binaries clean-dir clean-test-dir clean-coverage

clean-binaries:
	@rm -rf ${GOBIN}/pellcored
	@rm -rf ${GOBIN}/pellclientd

clean-dir:
	@rm -rf ~/.pellcored
	@rm -rf ~/.pellcore

all: install

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

###############################################################################
###                             Test commands                               ###
###############################################################################

run-test:
	@go test ${TEST_BUILD_FLAGS} ${TEST_DIR}

test :clean-test-dir run-test

test-hsm:
	@go test ${HSM_BUILD_FLAGS} ${TEST_DIR}

# Generate the test coverage
# "|| exit 1" is used to return a non-zero exit code if the tests fail
test-coverage:
	@go test ${TEST_BUILD_FLAGS} -coverprofile coverage.out ${TEST_DIR} || exit 1

coverage-report: test-coverage
	@go tool cover -html=coverage.out -o coverage.html

clean-coverage:
	@rm -f coverage.out
	@rm -f coverage.html

clean-test-dir:
	@rm -rf x/xmsg/client/integrationtests/.pellcored
	@rm -rf x/xmsg/client/querytests/.pellcored
	@rm -rf x/relayer/client/querytests/.pellored

###############################################################################
###                          Install commands                               ###
###############################################################################

build-testnet-ubuntu: go.sum
		docker build -t pellcore-ubuntu --platform linux/amd64 -f ./Dockerfile-ignite3-ubuntu .
		docker create --name temp-container pellcore-ubuntu
		docker cp temp-container:/go/bin/pellclientd .
		docker cp temp-container:/go/bin/pellcored .
		docker rm temp-container

install: go.sum
		@echo "--> Installing pellcored & pellclientd"
		@CGO_CFLAGS="-Wno-deprecated-declarations" go install -mod=readonly $(BUILD_FLAGS) ./cmd/pellcored
		@CGO_CFLAGS="-Wno-deprecated-declarations" go install -mod=readonly $(BUILD_FLAGS) ./cmd/pellclientd
install-pellclient: go.sum
		@echo "--> Installing pellclientd"
		@go install -mod=readonly $(BUILD_FLAGS) ./cmd/pellclientd

install-pellcore: go.sum
		@echo "--> Installing pellcored"
		@go install -mod=readonly $(BUILD_FLAGS) ./cmd/pellcored
		
install-e2e: go.sum
		@echo "--> Installing pelle2e"
		@CGO_CFLAGS="-Wno-deprecated-declarations" go install -mod=readonly $(BUILD_FLAGS) ./cmd/pelle2e

# running with race detector on will be slow
install-pellclient-race-test-only-build: go.sum
		@echo "--> Installing pellclientd"
		@go install -race -mod=readonly $(BUILD_FLAGS) ./cmd/pellclientd

install-pelltool: go.sum
		@echo "--> Installing pelltool"
		@go install -mod=readonly $(BUILD_FLAGS) ./cmd/pelltool

###############################################################################
###                             Local network                               ###
###############################################################################

init:
	./standalone-network/init.sh

run:
	./standalone-network/run.sh

chain-init: clean install-pellcore init
chain-run: clean install-pellcore init run
chain-stop:
	@killall pellcored
	@killall tail

test-xmsg:
	./standalone-network/xmsg-creator.sh

###############################################################################
###                                 Linting            	                    ###
###############################################################################

lint-pre:
	@test -z $(gofmt -l .)
	@GOFLAGS=$(GOFLAGS) go mod verify

lint: lint-pre
	@golangci-lint run

lint-cosmos-gosec:
	@bash ./scripts/cosmos-gosec.sh

gosec:
	gosec  -exclude-dir=localnet ./...

###############################################################################
###                           Generation commands  		                    ###
###############################################################################

proto:
	@echo "--> Removing old Go types "
	@find . -name '*.pb.go' -type f -delete
	@echo "--> Generating new Go types from protocol buffer files"
	@bash ./scripts/protoc-gen-go.sh
	@buf format -w
.PHONY: proto

typescript:
	@echo "--> Generating TypeScript bindings"
	@bash ./scripts/protoc-gen-typescript.sh
.PHONY: typescript

proto-format:
	@bash ./scripts/proto-format.sh

openapi:
	@echo "--> Generating OpenAPI specs"
	@bash ./scripts/protoc-gen-openapi.sh
.PHONY: openapi

specs:
	@echo "--> Generating module documentation"
	@go run ./scripts/gen-spec.go
.PHONY: specs

docs-pellcored:
	@echo "--> Generating pellcored documentation"
	@bash ./scripts/gen-docs-pellcored.sh
.PHONY: docs-pellcored

mocks:
	@echo "--> Generating mocks"
	@bash ./scripts/mocks-generate.sh
.PHONY: mocks

generate: proto openapi specs typescript docs-pellcored
.PHONY: generate

###############################################################################
###                         E2E tests and localnet                          ###
###############################################################################

pellnode:
	@echo "Building pellnode"
	$(DOCKER) build -t pellnode -f ./Dockerfile-localnet .
	$(DOCKER) build -t orchestrator -f contrib/localnet/orchestrator/Dockerfile.fastbuild .
.PHONY: pellnode

install-pelle2e: go.sum
	@echo "--> Installing pelle2e"
	@go install -mod=readonly ./cmd/pelle2e
.PHONY: install-pelle2e

start-e2e-test: pellnode
	@echo "--> Starting e2e test"
	cd contrib/localnet/ && $(DOCKER) compose up -d

start-e2e-admin-test: pellnode
	@echo "--> Starting e2e admin test"
	cd contrib/localnet/ && $(DOCKER) compose -f docker-compose.yml -f docker-compose-admin.yml up -d

start-e2e-import-mainnet-test: pellnode
	@echo "--> Starting e2e import-data test"
	cd contrib/localnet/  && ./scripts/import-data.sh mainnet && $(DOCKER) compose -f docker-compose.yml -f docker-compose-import-data.yml up -d

start-e2e-performance-test: pellnode
	@echo "--> Starting e2e performance test"
	cd contrib/localnet/ && $(DOCKER) compose -f docker-compose.yml -f docker-compose-performance.yml up -d


start-stress-test: pellnode
	@echo "--> Starting stress test"
	cd contrib/localnet/ && $(DOCKER) compose -f docker-compose.yml -f docker-compose-stresstest.yml up -d

start-upgrade-test:
	@echo "--> Starting upgrade test"
	$(DOCKER) build -t pellnode -f ./Dockerfile-upgrade .
	$(DOCKER) build -t orchestrator -f contrib/localnet/orchestrator/Dockerfile.fastbuild .
	cd contrib/localnet/ && $(DOCKER) compose -f docker-compose.yml -f docker-compose-upgrade.yml up -d

start-upgrade-test-light:
	@echo "--> Starting light upgrade test (no PellChain state populating before upgrade)"
	$(DOCKER) build -t pellnode -f ./Dockerfile-upgrade .
	$(DOCKER) build -t orchestrator -f contrib/localnet/orchestrator/Dockerfile.fastbuild .
	cd contrib/localnet/ && $(DOCKER) compose -f docker-compose.yml -f docker-compose-upgrade-light.yml up -d
build-localnet:
	@echo "--> Starting build localnet docker images"
	cd contrib/localnet/ && \
	docker compose -f docker-compose.build.yml build pellnode && \
	docker compose -f docker-compose.build.yml build pell-contracts && \
	docker compose -f docker-compose.build.yml build orchestrator && \
	docker compose -f docker-compose.build.yml build eth
	
start-localnet:
	@echo "--> Starting localnet"
	cd contrib/localnet/ && \
	docker compose -p pell-e2e -f docker-compose.e2e.yml down --volumes && \
	docker compose -p pell-e2e -f docker-compose.e2e.yml up  -d && \
	docker compose -p pell-e2e -f docker-compose.e2e.yml run --rm orchestrator /work/start-e2e-local.sh

start-local-e2e: build-localnet start-localnet

stop-test:
	cd contrib/localnet/ && $(DOCKER) compose down --remove-orphans

###############################################################################
###                              Monitoring                                 ###
###############################################################################

start-monitoring:
	@echo "Starting monitoring services"
	cd contrib/localnet/grafana/ && ./get-tss-address.sh
	cd contrib/localnet/ && $(DOCKER) compose -f docker-compose-monitoring.yml up -d

stop-monitoring:
	@echo "Stopping monitoring services"
	cd contrib/localnet/ && $(DOCKER) compose -f docker-compose-monitoring.yml down

###############################################################################
###                                GoReleaser  		                        ###
###############################################################################

PACKAGE_NAME          := github.com/pell-chain/node
GOLANG_CROSS_VERSION  ?= v1.23
GOPATH ?= '$(HOME)/go'
release-dry-run:
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-e "GITHUB_TOKEN=${GITHUB_TOKEN}" \
		-e GOPRIVATE="github.com/0xPellNetwork" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-v ${GOPATH}/pkg:/go/pkg \
		-w /go/src/$(PACKAGE_NAME) \
    --entrypoint sh \
    ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
    -c "\
    git config --global url.https://${GITHUB_TOKEN}:x-oauth-basic@github.com/0xPellNetwork.insteadOf https://github.com/0xPellNetwork && \
    goreleaser --clean --skip=publish --snapshot"

release:
	@if [ ! -f ".release-env" ]; then \
		echo "\033[91m.release-env is required for release\033[0m";\
		exit 1;\
	fi
	docker run \
		--rm \
		--privileged \
		-e CGO_ENABLED=1 \
		-e "GITHUB_TOKEN=${GITHUB_TOKEN}" \
		-e GOPRIVATE="github.com/0xPellNetwork" \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v `pwd`:/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
    --entrypoint sh \
    ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
    -c "\
    git config --global url.https://${GITHUB_TOKEN}:x-oauth-basic@github.com/0xPellNetwork.insteadOf https://github.com/0xPellNetwork && \
    goreleaser release --clean --skip=validate"

###############################################################################
###                     Local Mainnet Development                           ###
###############################################################################

mainnet-pellrpc-node:
	cd contrib/mainnet/pellcored && DOCKER_TAG=$(DOCKER_TAG) docker-compose up

mainnet-bitcoind-node:
	cd contrib/mainnet/bitcoind && DOCKER_TAG=$(DOCKER_TAG) docker-compose up

ignite3-pellrpc-node:
	cd contrib/ignite3/pellcored && DOCKER_TAG=$(DOCKER_TAG) docker-compose up

###############################################################################
###                               Debug Tools                               ###
###############################################################################

filter-missed-btc: install-pelltool
	pelltool filterdeposit btc --config ./tool/filter_missed_deposits/pelltool_config.json

filter-missed-eth: install-pelltool
	pelltool filterdeposit eth \
		--config ./tool/filter_missed_deposits/pelltool_config.json \
		--evm-max-range 1000 \
		--evm-start-block 19464041

.PHONY: fmt

lint-imports:
	@find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" | while read -r file; do \
		goimports-reviser -company-prefixes github.com/0xPellNetwork/aegis -rm-unused -format "$$file"; \
	done
