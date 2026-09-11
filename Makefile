.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: build
build: ## Build the lsgo binary.
	@go build -v -o bin/lsgo ./cmd/lsgo/...

.PHONY: test
test: ## Run the tests.
	@go test ./...

.PHONY: lint
lint: ## Run the linter.
	@golangci-lint run --fix

.PHONY: fmt
fmt: ## Format the code.
	@golangci-lint fmt

.PHONY: check
check: fmt lint test ## Helper to format, lint and test in one go

.PHONY: clean
clean: ## Clean the build artifacts.
	rm -rf bin/lsgo