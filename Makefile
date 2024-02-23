.SHELL := /usr/bin/env bash

# Environment
GO := $(shell which go)
DOCKER := $(shell which docker)
GOTEST = $(GO) test
GOLIST := $(shell $(GO) list ./... | grep -v /vendor/)
PWD := $(shell pwd)
COVERAGE_FILE = profile.cov
MKDOCS := "${PWD}/.venv/bin/mkdocs"


# Fancy colors
GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

## Testing

test: ## Run all tests
	$(GOTEST) -race -v $(GOLIST)

coverage: ## Run tests with coverage and export it into 'profile.cov'.
	$(GOTEST) -cover -covermode=count -coverprofile=$(COVERAGE_FILE) ./...
	$(GO) tool cover -func $(COVERAGE_FILE)

show-coverage: coverage ## Run coverage and open the rendered coverage site in the browser
	$(GO) tool cover -html=$(COVERAGE_FILE)
	@echo -e "\n=> Coverage report opened in your default browser"

## Lint

lint: lint-go lint-yaml ## Run all linters

lint-go: ## Lint all GO files
	$(DOCKER) run --rm -it -v $(PWD):/app -w /app $(GOLANGCI_LINT_IMAGE) golangci-lint run --deadline=65s

lint-yaml: ## Lint all YAML files
	$(DOCKER) run --rm -it -v $(PWD):/data $(YAMLLINT_IMAGE) -f parsable $(YAMLFILES)

## Docs

serve-docs:
	@cd docs && $(MKDOCS) serve -a localhost:8080

serve-godocs:
	@pkgsite -open


## Examples
example-terraform:
	go run examples/terraform/main.go \
		-data examples/terraform/inventory \
		-templates examples/terraform/templates \
		-output examples/terraform/compiled \
		-target dev

example-terraform-azure:
	go run examples/terraform_azure/main.go \
		-data examples/terraform_azure/inventory \
		-templates examples/terraform_azure/templates \
		-output examples/terraform_azure/compiled \
		-target develop

example-terraform-azure_debug:
	dlv debug examples/terraform_azure/main.go -- \
		-data examples/terraform_azure/inventory \
		-templates examples/terraform_azure/templates \
		-output examples/terraform_azure/compiled \
		-target develop

example-keyvault:
	go run examples/keyvault/main.go \
		-data examples/keyvault/inventory \
		-templates examples/keyvault/templates \
		-output examples/keyvault/compiled \
		-target develop

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

