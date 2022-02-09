BINARY=packer-plugin-upcloud
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)
COUNT?=1
TEST?=$(shell go list ./...)
GOBIN=${GOPATH}/bin
PACKER_SDC=$(GOBIN)/packer-sdc
default: build

test:
	@go test -race -count $(COUNT) $(TEST) -timeout=3m

test_integration: build
	cp ${BINARY} builder/upcloud/
	PACKER_ACC=1 go test -count 1 -v ./...  -timeout=120m

lint:
	go vet .
	golint .

build:
	go build -v -o ${BINARY}

install: build
	@mkdir -p ~/.packer.d/plugins
	install ${BINARY} ~/.packer.d/plugins/

install-packer-sdc: ## Install packer sofware development command
	go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@${HASHICORP_PACKER_PLUGIN_SDK_VERSION}

ci-release-docs: install-packer-sdc
	@$(PACKER_SDC) renderdocs -src docs -partials docs-partials/ -dst docs/
	@/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

plugin-check: install-packer-sdc build
	$(PACKER_SDC) plugin-check ${BINARY}

generate: install-packer-sdc
	@go generate ./...
	packer-sdc renderdocs -src ./docs -dst ./.docs -partials ./docs-partials

.PHONY: default test test_integration lint build install
