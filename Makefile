SHELL := /bin/sh

BINARY := rollbar-cli
BUILD_DIR := bin
PKG := .
SKILL_NAME := rollbar-cli
SKILL_SOURCE_DIR := .ai/skills/$(SKILL_NAME)
AI_SKILL_DIRS ?= $(HOME)/.codex/skills $(HOME)/.claude/skills $(HOME)/.config/claude/skills $(HOME)/.cursor/skills $(HOME)/.windsurf/skills

.PHONY: all build install install-skill test test-cover clean help

all: build

help:
	@echo "Targets:"
	@echo "  make build       Build $(BINARY) into $(BUILD_DIR)/"
	@echo "  make install     Install $(BINARY) with 'go install'"
	@echo "  make install-skill  Install .ai skill into common AI tool skill dirs"
	@echo "  make test        Run unit tests"
	@echo "  make test-cover  Run unit tests with coverage"
	@echo "  make clean       Remove build artifacts"

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(PKG)

install: install-skill
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

test:
	go test ./...

test-cover:
	go test ./... -cover

clean:
	rm -rf $(BUILD_DIR)
