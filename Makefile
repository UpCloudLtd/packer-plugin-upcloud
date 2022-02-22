BINARY=packer-plugin-upcloud
HASHICORP_PACKER_PLUGIN_SDK_VERSION?=$(shell go list -m github.com/hashicorp/packer-plugin-sdk | cut -d " " -f2)
COUNT?=1
TEST?=$(shell go list ./...)

ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

PACKER_SDC=$(GOBIN)/packer-sdc
PACKER_SDC_RENDER_DOCS=$(PACKER_SDC) renderdocs -src docs-src/ -partials docs-partials/ -dst docs/


default: build

test:
	@go test -race -count $(COUNT) $(TEST) -timeout=3m

test_integration: build
	cp $(BINARY) builder/upcloud/
	PACKER_ACC=1 go test -count 1 -v ./...  -timeout=120m

lint:
	go vet .
	golint .

build:
	go build -v

install: build
	@mkdir -p ~/.packer.d/plugins
	install $(BINARY) ~/.packer.d/plugins/

install-packer-sdc: ## Install packer sofware development command
	go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@$(HASHICORP_PACKER_PLUGIN_SDK_VERSION)

ci-release-docs: install-packer-sdc generate
	@/bin/sh -c "[ -d docs ] && zip -r docs.zip docs/"

plugin-check: install-packer-sdc build
	$(PACKER_SDC) plugin-check $(BINARY)

generate: fmt install-packer-sdc
	@PATH=$(PATH):$(GOBIN) go generate ./...
	@rm -fr $(CURDIR)/docs # renderdocs doesn't seem to properly overwrite files
	$(PACKER_SDC_RENDER_DOCS)

fmt:
	packer fmt example/
	packer fmt -recursive docs-partials/

.PHONY: default test test_integration lint build install
