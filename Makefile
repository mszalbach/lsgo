.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: dist
dist: ## Create a test release.
	@goreleaser release --snapshot --clean

.PHONY: build
build: ## Only executable and Docker, without the rest.
	@goreleaser release --snapshot --clean --skip=archive,sbom

.PHONY: test
test: ## Run the tests.
	@go test ./...

.PHONY: e2e
e2e: build ## Run the e2e tests.
	@cd e2e && go test ./...

.PHONY: lint
lint: ## Run the linter.
	@golangci-lint run --fix
	@cd e2e && golangci-lint run --config ../.golangci.yaml ./... --fix

.PHONY: lint-ci
lint-ci: ## Run the linter for ci without auto fix.
	@golangci-lint run
	@cd e2e && golangci-lint run --config ../.golangci.yaml ./...

.PHONY: fmt
fmt: ## Format the code.
	@golangci-lint fmt
	@cd e2e && golangci-lint fmt

.PHONY: check
check: fmt lint test ## Helper to format, lint and test in one go.

.PHONY: clean
clean: ## Clean the build artifacts.
	rm -rf bin/lsgo