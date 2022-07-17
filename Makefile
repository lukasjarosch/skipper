.SHELL := /usr/bin/env bash

# Environment
GO := $(shell which go)
DOCKER := $(shell which docker)
GOTEST = $(GO) test
GOLIST := $(shell $(GO) list ./... | grep -v /vendor/)
PWD := $(shell pwd)

# Fancy colors
GREEN  := $(shell tput -Txterm setaf 2)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

## Examples
example-terraform-dev:
	go run examples/terraform/main.go \
		-data examples/terraform/inventory \
		-templates examples/terraform/templates \
		-output examples/terraform/compiled \
		-target dev


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

