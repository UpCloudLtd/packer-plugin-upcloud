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
PACKER_SDC_RENDER_DOCS=$(PACKER_SDC) renderdocs -src docs-src/ -partials docs-partials/ -dst .docs/


default: build

test:
	@go test -race -count $(COUNT) $(TEST) -timeout=3m

test_integration: build
	cp $(BINARY) builder/upcloud/
	PACKER_ACC=1 go test -count 1 -v $(TESTARGS) ./...  -timeout=120m

build:
	go build -v

install: build
	@mkdir -p ~/.packer.d/plugins
	install $(BINARY) ~/.packer.d/plugins/

install-packer-sdc: ## Install packer sofware development command
	go install github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@$(HASHICORP_PACKER_PLUGIN_SDK_VERSION)

plugin-check: install-packer-sdc build
	$(PACKER_SDC) plugin-check $(BINARY)

docs: fmt install-packer-sdc
	@PATH=$(PATH):$(GOBIN) go generate ./...
	@if [ -d ".docs" ]; then rm -r ".docs"; fi
	$(PACKER_SDC_RENDER_DOCS)
	@./.web-docs/scripts/compile-to-webdocs.sh "." ".docs" ".web-docs" "UpCloudLtd"
	@rm -r ".docs"

fmt:
	packer fmt builder/upcloud/test-fixtures/hcl2
	packer fmt example/
	packer fmt -recursive docs-partials/

clean:
	find . -name "packer_log_*" -delete
	find . -name "TestBuilderAcc_*" -delete
	find . -name "packer-plugin-upcloud" -delete

.PHONY: default test test_integration lint build install docs
