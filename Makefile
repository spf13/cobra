BIN="./bin"
SRC=$(shell find . -name "*.go")
GOPATH=$(shell go env GOPATH)

ifeq (, $(shell which golangci-lint))
$(warning "could not find golangci-lint in PATH, run: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.50.0")
endif

ifeq (, $(shell which richgo))
$(warning "could not find richgo in PATH, run: go install github.com/kyoh86/richgo@latest")
endif

.PHONY: fmt lint test install_deps clean

default: all

all: fmt test

fmt:
	$(info ******************** checking formatting ********************)
	@test -z $(shell gofmt -l $(SRC)) || (gofmt -d $(SRC); exit 1)

lint:
	$(info ******************** running lint tools ********************)
	golangci-lint run -v

test: install_deps
	$(info ******************** running tests ********************)
	richgo test -v ./...

install_deps:
	$(info ******************** downloading dependencies ********************)
	go get -v ./...

clean:
	rm -rf $(BIN)
