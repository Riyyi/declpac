.PHONY: build fmt fmtcheck vet check githook install uninstall

PKG := github.com/Riyyi/declpac
BIN := declpac
BUILD_DIR := bin
BUILD := $(BUILD_DIR)/$(BIN)
PREFIX ?= /usr/local
DESTDIR ?=

STAGED_GO := $(shell git diff --cached --name-only --diff-filter=ACM | grep '\.go$$')
STAGED_PKGS := $(shell echo "$(STAGED_GO)" | tr ' ' '\n' | xargs -I{} dirname {} | sort -u | sed 's|^|./|')

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD) ./cmd/declpac

fmt:
	@gofmt -w cmd pkg

fmtcheck:
	@if [ -z "$(STAGED_GO)" ]; then exit 0; fi; \
	unformatted=$$(gofmt -l $(STAGED_GO)); \
	if [ -n "$$unformatted" ]; then \
		echo "$$unformatted: unformatted staged file"; \
		exit 1; \
	fi

vet:
	@if [ -z "$(STAGED_GO)" ]; then exit 0; fi; \
	go vet $(STAGED_PKGS)

check: fmtcheck vet

githook:
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make check' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Installed pre-commit hook: .git/hooks/pre-commit"

install: build
	@mkdir -p $(DESTDIR)$(PREFIX)/bin
	@install -m 0755 $(BUILD) $(DESTDIR)$(PREFIX)/bin/$(BIN)

uninstall:
	@rm -f $(DESTDIR)$(PREFIX)/bin/$(BIN)
