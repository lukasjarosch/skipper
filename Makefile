.SHELL := /usr/bin/env bash

# Environment
GO := $(shell which go)
DOCKER := $(shell which docker)
GOTEST = $(GO) test
GOLIST := $(shell $(GO) list ./... | grep -v /vendor/)
PWD := $(shell pwd)
COVERAGE_FILE = profile.cov


# Fancy colors
GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

all: help

## Testing

test: ## Run all tests
	$(GOTEST) -race -v $(GOLIST)

coverage: ## Run tests with coverage and export it into 'profile.cov'. If 'COVERAGE_EXPORT' is true, 'COVERAGE_FILE' is written
	$(GOTEST) -cover -covermode=count -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func $(COVERAGE_FILE)
ifeq ($(COVERAGE_EXPORT), false)
	@rm $(COVERAGE_FILE)
endif

## Lint

lint: lint-go lint-yaml ## Run all linters

lint-go: ## Lint all GO files
	$(DOCKER) run --rm -it -v $(PWD):/app -w /app $(GOLANGCI_LINT_IMAGE) golangci-lint run --deadline=65s

lint-yaml: ## Lint all YAML files
	$(DOCKER) run --rm -it -v $(PWD):/data $(YAMLLINT_IMAGE) -f parsable $(YAMLFILES)



## Examples
example-terraform-dev:
	go run examples/terraform/main.go \
		-data examples/terraform/inventory \
		-templates examples/terraform/templates \
		-output examples/terraform/compiled \
		-target dev

example-external-classes:
	cd examples/external_classes && go run main.go

## Help

help: ## Show this help.
	@echo '==============================='
	@echo " ${GREEN}Skipper - Makefile${RESET}"
	@echo '==============================='
	@echo ''
	@echo 'Usage:'
	@echo '  make ${CYAN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${CYAN}%-20s${WHITE}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${WHITE}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

