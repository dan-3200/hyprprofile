BINARY_NAME ?= hyprprofile
LAUNCHER_NAME ?= hyprprofile-session-launcher
GUI_NAME ?= hyprprofile-gui
BUILD_DIR ?= bin
MAIN_PACKAGE ?= ./cmd/hyprprofile
LAUNCHER_PACKAGE ?= ./cmd/hyprprofile-session-launcher
GUI_PACKAGE ?= ./cmd/hyprprofile-gui
GO ?= go
PROJECT_CACHE_DIR := $(CURDIR)/.cache
TOOLS_DIR ?= $(CURDIR)/.tools/bin
GOLANGCI_LINT_VERSION ?= v2.15.0
GOLANGCI_LINT ?= $(TOOLS_DIR)/golangci-lint
GO_ENV := GOCACHE=$(PROJECT_CACHE_DIR)/go-build GOMODCACHE=$(PROJECT_CACHE_DIR)/go-mod
LINT_ENV := $(GO_ENV) GOLANGCI_LINT_CACHE=$(PROJECT_CACHE_DIR)/golangci-lint
# The current gotk4 generator declares free without including <stdlib.h> in one
# generated C wrapper. It is safe, but GCC emits a warning while compiling it.
GUI_CGO_CFLAGS ?= -Wno-builtin-declaration-mismatch

.DEFAULT_GOAL := help

.PHONY: help all check build build-gui run install test test-race coverage lint fmt tidy tools clean check-gui-deps

help: ## Show the available targets
	@awk 'BEGIN {FS = ":.*## "; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z_-]+:.*## / {printf "  %-12s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: check build ## Check and build the project

check: lint test ## Run all quality checks

build: ## Build the CLI in the build directory
	@mkdir -p $(BUILD_DIR)
	$(GO_ENV) $(GO) build -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	$(GO_ENV) $(GO) build -trimpath -o $(BUILD_DIR)/$(LAUNCHER_NAME) $(LAUNCHER_PACKAGE)

build-gui: check-gui-deps ## Build the native GTK 4 and libadwaita GUI
	@mkdir -p $(BUILD_DIR)
	$(GO_ENV) CGO_CFLAGS="$(CGO_CFLAGS) $(GUI_CGO_CFLAGS)" $(GO) build -tags gui -trimpath -o $(BUILD_DIR)/$(GUI_NAME) $(GUI_PACKAGE)

run: ## Run the CLI; pass arguments with ARGS="..."
	$(GO_ENV) $(GO) run $(MAIN_PACKAGE) $(ARGS)

install: ## Install the CLI and session launcher using the active Go environment
	$(GO_ENV) $(GO) install $(MAIN_PACKAGE)
	$(GO_ENV) $(GO) install $(LAUNCHER_PACKAGE)

test: ## Run the test suite
	$(GO_ENV) $(GO) test ./...

test-race: ## Run tests with the race detector
	$(GO_ENV) $(GO) test -race ./...

coverage: ## Generate coverage.out and print the coverage summary
	$(GO_ENV) $(GO) test -coverprofile=coverage.out ./...
	$(GO_ENV) $(GO) tool cover -func=coverage.out

lint: check-golangci-lint ## Run the configured linters
	$(LINT_ENV) $(GOLANGCI_LINT) run ./...

fmt: check-golangci-lint ## Format Go source files and imports
	$(LINT_ENV) $(GOLANGCI_LINT) fmt ./...

tidy: ## Normalize go.mod and go.sum
	$(GO_ENV) $(GO) mod tidy

tools: ## Install pinned development tools locally
	@mkdir -p $(TOOLS_DIR)
	$(GO_ENV) GOBIN=$(TOOLS_DIR) $(GO) install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

clean: ## Remove generated build and coverage artifacts
	$(GO_ENV) $(GO) clean
	rm -rf -- $(BUILD_DIR)
	rm -rf -- $(PROJECT_CACHE_DIR)
	rm -f -- coverage.out

.PHONY: check-golangci-lint
check-golangci-lint:
	@test -x $(GOLANGCI_LINT) || { \
		echo "golangci-lint is required; run 'make tools'"; \
		exit 1; \
	}

check-gui-deps:
	@pkg-config --exists gtk4 libadwaita-1 gobject-introspection-1.0 appstream || { \
		echo "GTK GUI dependencies are required; on Fedora run: sudo dnf install gtk4-devel libadwaita-devel gobject-introspection-devel appstream-devel"; \
		exit 1; \
	}
