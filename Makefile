SHELL := /bin/sh

BINARY := rollbar-cli
BUILD_DIR := bin
PKG := .
SKILL_NAME := rollbar-cli
SKILL_SOURCE_DIR := .ai/skills/$(SKILL_NAME)
AI_SKILL_DIRS ?= $(HOME)/.codex/skills $(HOME)/.claude/skills $(HOME)/.config/claude/skills $(HOME)/.cursor/skills $(HOME)/.windsurf/skills
CROSS_PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: all build build-cross install install-skill install-all test test-cover clean help

all: build

help:
	@echo "Targets:"
	@printf "  %-18s %s\n" "make build" "Build $(BINARY) into $(BUILD_DIR)/"
	@printf "  %-18s %s\n" "make build-cross" "Build common cross-platform binaries into $(BUILD_DIR)/"
	@printf "  %-18s %s\n" "make install" "Install $(BINARY) with 'go install'"
	@printf "  %-18s %s\n" "make install-skill" "Install .ai skill into common AI tool skill dirs"
	@printf "  %-18s %s\n" "make install-all" "Install $(BINARY) and the .ai skill"
	@printf "  %-18s %s\n" "make test" "Run unit tests"
	@printf "  %-18s %s\n" "make test-cover" "Run unit tests with coverage"
	@printf "  %-18s %s\n" "make clean" "Remove build artifacts"

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(PKG)

build-cross:
	@mkdir -p $(BUILD_DIR)
	@set -e; \
	for target in $(CROSS_PLATFORMS); do \
		goos=$${target%/*}; \
		goarch=$${target#*/}; \
		output="$(BUILD_DIR)/$(BINARY)_$${goos}_$${goarch}"; \
		if [ "$$goos" = "windows" ]; then \
			output="$$output.exe"; \
		fi; \
		echo "Building $$output"; \
		CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch go build -trimpath -ldflags="-s -w" -o "$$output" $(PKG); \
	done

install:
	go install $(PKG)

install-skill:
	@set -e; \
	for dir in $(AI_SKILL_DIRS); do \
		target="$$dir/$(SKILL_NAME)"; \
		mkdir -p "$$dir"; \
		rm -rf "$$target"; \
		cp -R "$(SKILL_SOURCE_DIR)" "$$target"; \
		echo "Installed $(SKILL_SOURCE_DIR) -> $$target"; \
	done

install-all: install install-skill

test:
	go test ./...

test-cover:
	go test ./... -cover

clean:
	rm -rf $(BUILD_DIR)
