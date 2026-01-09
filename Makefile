BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')
APPNAME := zigchain

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
TM_VERSION := $(shell go list -m github.com/cometbft/cometbft | sed 's:.* ::')
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build

GO_SYSTEM_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1)
REQUIRE_GO_VERSION = 1.25.4

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  build_tags += ledger
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace := $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=$(APPNAME) \
          -X github.com/cosmos/cosmos-sdk/version.AppName=$(APPNAME)d \
          -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
          -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
          -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
          -X github.com/cometbft/cometbft/version.TMCoreSemVer=$(TM_VERSION)^

ldflags += -w -s
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

##############
###  Test  ###
##############

test-unit:
	@echo Running unit tests...
	@go test -mod=readonly -v -timeout 30m ./...

test-race:
	@echo Running unit tests with race condition reporting...
	@go test -mod=readonly -v -race -timeout 30m ./...

test-cover:
	@echo Running unit tests and creating coverage report...
	@go test -mod=readonly -v -timeout 30m -coverprofile=$(COVER_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVER_FILE) -o $(COVER_HTML_FILE)
	@rm $(COVER_FILE)

bench:
	@echo Running unit tests with benchmarking...
	@go test -mod=readonly -v -timeout 30m -bench=. ./...

test: check_version test-unit govet govulncheck

.PHONY: test test-unit test-race test-cover bench

#################
###  Install  ###
#################

check_version:
	@echo "Checking Go version: $(GO_SYSTEM_VERSION) (required: same MAJOR.MINOR as $(REQUIRE_GO_VERSION))"
	@REQUIRED_MAJOR_MINOR=$$(echo "$(REQUIRE_GO_VERSION)" | cut -d. -f1,2); \
	SYSTEM_MAJOR_MINOR=$$(echo "$(GO_SYSTEM_VERSION)" | cut -d. -f1,2); \
	if [ "$$REQUIRED_MAJOR_MINOR" != "$$SYSTEM_MAJOR_MINOR" ]; then \
		echo "ERROR: Go version with MAJOR.MINOR $(REQUIRE_GO_VERSION) is required for $(VERSION) of zigchain."; \
		echo "Required MAJOR.MINOR: $$REQUIRED_MAJOR_MINOR"; \
		echo "Current version: $(GO_SYSTEM_VERSION) (MAJOR.MINOR: $$SYSTEM_MAJOR_MINOR)"; \
		exit 1; \
	fi; \
	if [ "$(shell printf '%s\n' "$(REQUIRE_GO_VERSION)" "$(GO_SYSTEM_VERSION)" | sort -V | head -n1)" != "$(REQUIRE_GO_VERSION)" ]; then \
		echo "ERROR: Go version $(REQUIRE_GO_VERSION) or higher is required for $(VERSION) of zigchain."; \
		echo "Required version: $(REQUIRE_GO_VERSION)"; \
		echo "Current version: $(GO_SYSTEM_VERSION)"; \
		exit 1; \
	fi

all: install lint test

build: BUILD_ARGS=-o $(BUILDDIR)/

build install: check_version go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./cmd/zigchaind

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	go mod verify
	go mod tidy
	@echo "--> Download go modules to local cache"
	go mod download

.PHONY: check_version all build install $(BUILDDIR)/ go.sum

##################
###  Protobuf  ###
##################

# Use this target if you do not want to use Ignite for generating proto files
GOLANG_PROTOBUF_VERSION=1.28.1
GRPC_GATEWAY_VERSION=1.16.0
GRPC_GATEWAY_PROTOC_GEN_OPENAPIV2_VERSION=2.20.0

proto-deps:
	@echo "Installing proto deps"
	@go install github.com/bufbuild/buf/cmd/buf@v1.50.0
	@go install github.com/cosmos/gogoproto/protoc-gen-gogo@latest
	@go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(GOLANG_PROTOBUF_VERSION)
	@go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v$(GRPC_GATEWAY_VERSION)
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v$(GRPC_GATEWAY_PROTOC_GEN_OPENAPIV2_VERSION)
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto-gen:
	@echo "Generating protobuf files..."
	@ignite generate proto-go --yes

.PHONY: proto-deps proto-gen

#################
###  Linting  ###
#################

golangci_lint_cmd=golangci-lint
golangci_version=v1.61.0

lint:
	@echo "--> Running linter"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run ./... --timeout 15m

lint-fix:
	@echo "--> Running linter and fixing issues"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(golangci_version)
	@$(golangci_lint_cmd) run ./... --fix --timeout 15m

.PHONY: lint lint-fix

###################
### Development ###
###################

govet:
	@echo Running go vet...
	@go list ./... | grep -v "^zigchain/api" | xargs go vet

govulncheck:
	@echo Running govulncheck...
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

.PHONY: govet govulncheck

###################
###  Security   ###
###################

gosec:
	@echo "Running gosec security scanner..."
	@go install github.com/securego/gosec/v2/cmd/gosec@latest
	@gosec -exclude-dir api -exclude-dir docs -exclude-generated -sort -color ./...

staticcheck:
	@echo "Running staticcheck (includes security checks)..."
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@if command -v bat >/dev/null 2>&1; then \
		staticcheck -f text ./... | grep -Ev ".pulsar.|pb.go|pb.gw.go" | bat -l yaml --theme=DarkNeon --style=plain; \
	else \
		staticcheck -f text ./... | grep -Ev ".pulsar.|pb.go|pb.gw.go"; \
	fi

# Check for common issues in protobuf files
protobuf-security:
	@echo "Running protobuf security checks..."
	@go install github.com/bufbuild/buf/cmd/buf@latest
	@buf lint
	@buf breaking --against '.git#branch=main'

# Check for common issues in dependencies
dep-security:
	@echo "Running dependency security checks..."
	@go install github.com/google/go-licenses@latest
	@go-licenses check ./...
	@go list -m all | grep -v "github.com/cosmos/cosmos-sdk" | grep -v "github.com/cometbft/cometbft" | xargs go list -m -versions

security-audit: gosec staticcheck govulncheck protobuf-security dep-security
	@echo "Complete security audit completed"

.PHONY: gosec staticcheck protobuf-security dep-security security-audit

###############################################################################
###                                Release                                  ###
###############################################################################

GO_VERSION := $(shell cat go.mod | grep -E 'go [0-9].[0-9]+' | cut -d ' ' -f 2)
GORELEASER_IMAGE := ghcr.io/goreleaser/goreleaser-cross:v$(REQUIRE_GO_VERSION)
COSMWASM_VERSION := $(shell go list -m github.com/CosmWasm/wasmvm/v2 | sed 's/.* //')

# create tag and run goreleaser without publishing
# errors are possible while running goreleaser - the process can run for >30 min
# if the build is failing due to timeouts use goreleaser-build-local instead
create-release-dry-run:
ifneq ($(strip $(TAG)),)
	@echo "--> Dry running release for tag: $(TAG)"
	@echo "--> Create tag: $(TAG) dry run"
	git tag -s $(TAG) -m $(TAG)
	git push origin $(TAG) --dry-run
	@echo "--> Delete local tag: $(TAG)"
	@git tag -d $(TAG)
	@echo "--> Running goreleaser"
	@go install github.com/goreleaser/goreleaser@latest
	@docker run \
		--rm \
		-e CGO_ENABLED=1 \
		-e TM_VERSION=$(TM_VERSION) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v `pwd`:/go/src/$(APPNAME)d \
		-w /go/src/$(APPNAME)d \
		$(GORELEASER_IMAGE) \
		release \
		--snapshot \
		--skip=publish \
		--verbose \
		--clean
	@rm -rf dist/
	@echo "--> Done create-release-dry-run for tag: $(TAG)"
else
	@echo "--> No tag specified, skipping tag release"
endif

# Build static binaries for linux/amd64 using docker buildx
# Pulled from neutron-org/neutron: https://github.com/neutron-org/neutron/blob/v4.2.2/Makefile#L107
build-static-linux-amd64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name $(APPNAME)builder || true
	$(DOCKER) buildx use $(APPNAME)builder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TAGS=$(build_tags_comma_sep),muslc \
		--platform linux/amd64 \
		-t $(APPNAME)d-static-amd64 \
		-f Dockerfile . \
		--load
	$(DOCKER) rm -f $(APPNAME)binary || true
	$(DOCKER) create -ti --name $(APPNAME)binary $(APPNAME)d-static-amd64
	$(DOCKER) cp $(APPNAME)binary:/usr/local/bin/ $(BUILDDIR)/$(APPNAME)d-linux-amd64
	$(DOCKER) rm -f $(APPNAME)binary

# Build static binaries for linux/arm64 using docker buildx
# Pulled from neutron-org/neutron: https://github.com/neutron-org/neutron/blob/v4.2.2/Makefile#L107
build-static-linux-arm64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name $(APPNAME)builder || true
	$(DOCKER) buildx use $(APPNAME)builder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg BUILD_TAGS=$(build_tags_comma_sep),muslc \
		--platform linux/arm64 \
		-t $(APPNAME)d-static-arm64 \
		-f Dockerfile . \
		--load
	$(DOCKER) rm -f $(APPNAME)binary || true
	$(DOCKER) create -ti --name $(APPNAME)binary $(APPNAME)d-static-arm64
	$(DOCKER) cp $(APPNAME)binary:/usr/local/bin/ $(BUILDDIR)/$(APPNAME)d-linux-arm64
	$(DOCKER) rm -f $(APPNAME)binary


# uses goreleaser to create static binaries for darwin on local machine
goreleaser-build-local: check_version
	docker run \
		--rm \
		-e CGO_ENABLED=1 \
		-e TM_VERSION=$(TM_VERSION) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-e VERSION=$(VERSION) \
		-v `pwd`:/go/src/$(APPNAME)d \
		-w /go/src/$(APPNAME)d \
		$(GORELEASER_IMAGE) \
		release \
		--snapshot \
		--skip=publish \
		--release-notes ./CHANGELOG.md \
		--timeout 90m \
		--verbose

# uses goreleaser to create static binaries for linux an darwin
# requires access to GITHUB_TOKEN which has to be available in the CI environment
ifdef GITHUB_TOKEN
ci-release:
	docker run \
		--rm \
		-e CGO_ENABLED=1 \
		-e GITHUB_TOKEN=$(GITHUB_TOKEN) \
		-e TM_VERSION=$(TM_VERSION) \
		-e COSMWASM_VERSION=$(COSMWASM_VERSION) \
		-v `pwd`:/go/src/$(APPNAME)d \
		-w /go/src/$(APPNAME)d \
		$(GORELEASER_IMAGE) \
		release \
		--release-notes ./CHANGELOG.md \
		--timeout=90m \
		--clean
else
ci-release:
	@echo "Error: GITHUB_TOKEN is not defined. Please define it before running 'make release'."
endif

# create tag and publish it
create-release:
ifneq ($(strip $(TAG)),)
	@echo "--> Running release for tag: $(TAG)"
	@echo "--> Create release tag: $(TAG)"
	git tag -s $(TAG) -m $(TAG)
	git push origin $(TAG)
	@echo "--> Done creating release tag: $(TAG)"
else
	@echo "--> No tag specified, skipping create-release"
endif

.PHONY: create-release-dry-run build-static-linux-amd64 build-static-linux-arm64 goreleaser-build-local ci-release create-release