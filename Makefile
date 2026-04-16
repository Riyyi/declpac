.PHONY: fmt vet check githook

STAGED_GO := $(shell git diff --cached --name-only --diff-filter=ACM | grep '\.go$$')
STAGED_PKGS := $(shell echo "$(STAGED_GO)" | tr ' ' '\n' | xargs -I{} dirname {} | sort -u | sed 's|^|./|')

fmt:
	@if [ -z "$(STAGED_GO)" ]; then exit 0; fi; \
	unformatted=$$(gofmt -l $(STAGED_GO)); \
	if [ -n "$$unformatted" ]; then \
		echo "$$unformatted: unformatted staged file"; \
		exit 1; \
	fi

vet:
	@if [ -z "$(STAGED_GO)" ]; then exit 0; fi; \
	go vet $(STAGED_PKGS)

check: fmt vet

githook:
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make check' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Installed pre-commit hook: .git/hooks/pre-commit"
