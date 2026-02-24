SHELL := /bin/sh

BINARY := rollbar-cli
BUILD_DIR := bin
PKG := .

.PHONY: all build install test test-cover clean help

all: build

help:
	@echo "Targets:"
	@echo "  make build       Build $(BINARY) into $(BUILD_DIR)/"
	@echo "  make install     Install $(BINARY) with 'go install'"
	@echo "  make test        Run unit tests"
	@echo "  make test-cover  Run unit tests with coverage"
	@echo "  make clean       Remove build artifacts"

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(PKG)

install:
	go install $(PKG)

test:
	go test ./...

test-cover:
	go test ./... -cover

clean:
	rm -rf $(BUILD_DIR)
