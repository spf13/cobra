
SHELL := /bin/bash

XC_OS := "linux darwin"
XC_ARCH := "amd64"
XC_PARALLEL := "2"
BIN := "../bin"
E2EPATH := "../_test"
SRC := $(shell find . -name "*.go")
OS := $(shell sh -c 'uname -s 2>/dev/null || echo not supported' | tr '[:upper:]' '[:lower:]')
ifeq ($(shell uname -m),x86_64)
ARCH ?= amd64
endif
ifeq ($(shell uname -m),i686)
ARCH ?= 386
endif
ifeq ($(shell uname -m),aarch64)
ARCH ?= arm
endif

cobracmd := $(BIN)/cobra_$(OS)_$(ARCH)

ifeq (, $(shell which gox))
$(warning "could not find gox in $(PATH), run: go get github.com/mitchellh/gox")
endif

.PHONY: all build

default: all

all: build

build: ## Builds cobra generator utility
	gox \
		-os=$(XC_OS) \
		-arch=$(XC_ARCH) \
		-parallel=$(XC_PARALLEL) \
		-output=$(BIN)/{{.Dir}}_{{.OS}}_{{.Arch}} \
		;

e2e: ## Creates fake project for e2e testing
	@echo "Creating fake project and adding commands for e2e testing..."
	@rm -rf $(E2EPATH)
	@mkdir -p $(E2EPATH)
	@cd $(E2EPATH) && \
		$(cobra) init \
			-a "Author Name <someauthor@somedomain.com>" \
			-l "MIT" \
			--pkg-name github.com/somecompany/cobragenerated && \
		$(cobra) add \
			-a "Author Name <someauthor@somedomain.com>" \
			-l "MIT" \
			-p rootCmd \
			version && \
		$(cobra) add \
			-a "Author Name <someauthor@somedomain.com>" \
			-l "MIT" \
			-p rootCmd \
			--shortdesc 'configuration subcommands' \
			--longdesc 'configuration subcommands. Use config to show or update \
configuration for this application.' \
			config  && \
		$(cobra) add \
			-a "Author Name <someauthor@somedomain.com>" \
			-l "MIT" \
			-p configCmd \
			--shortdesc 'show app configuration' \
			--longdesc 'show app configuration. Can be used to generate a default \
configuration file by piping its output to a file.' \
			show
