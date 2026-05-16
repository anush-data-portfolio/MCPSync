# =============================================================================
# MCPSync — Makefile
# =============================================================================

BINARY     := mcpsync
MODULE     := github.com/anush-data-portfolio/MCPSync
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS    := -s -w \
              -X $(MODULE)/cmd.version=$(VERSION) \
              -X $(MODULE)/cmd.commit=$(COMMIT) \
              -X $(MODULE)/cmd.buildTime=$(BUILD_TIME)

OUTPUT_DIR  := dist
INSTALL_DIR := /usr/local/bin
GOBIN       := $(shell go env GOPATH)/bin

.DEFAULT_GOAL := help

# Detect current platform for install-user
GOOS_LOCAL   := $(shell go env GOOS)
GOARCH_LOCAL := $(shell go env GOARCH)

# Colors (no-op if terminal doesn't support them)
GREEN  := \033[0;32m
YELLOW := \033[0;33m
CYAN   := \033[0;36m
RESET  := \033[0m

# ─────────────────────────────────────────────────────────────────────────────
##@ Development
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: build
build: ## Build binary for the current platform
	@printf "$(GREEN)Building $(BINARY) $(VERSION)…$(RESET)\n"
	@mkdir -p $(OUTPUT_DIR)
	@CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY) .
	@printf "$(GREEN)→ $(OUTPUT_DIR)/$(BINARY)$(RESET)\n"

.PHONY: run
run: build ## Build and run (pass sub-commands via ARGS="now")
	@$(OUTPUT_DIR)/$(BINARY) $(ARGS)

.PHONY: fmt
fmt: ## Format all Go source files
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@go vet ./...

.PHONY: tidy
tidy: ## Tidy go module graph
	@go mod tidy

.PHONY: test
test: ## Run tests
	@go test ./... -v -count=1

.PHONY: test-race
test-race: ## Run tests with the race detector
	@go test -race ./... -count=1

.PHONY: test-cover
test-cover: ## Run tests and output coverage report
	@go test ./... -coverprofile=coverage.out -count=1
	@go tool cover -func=coverage.out

# ─────────────────────────────────────────────────────────────────────────────
##@ Quality & Security
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: lint
lint: ## Run golangci-lint (install via: make setup)
	@which golangci-lint > /dev/null 2>&1 || \
		(printf "$(YELLOW)golangci-lint not found — run: make setup$(RESET)\n" && exit 1)
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with --fix
	@which golangci-lint > /dev/null 2>&1 || \
		(printf "$(YELLOW)golangci-lint not found — run: make setup$(RESET)\n" && exit 1)
	@golangci-lint run --fix ./...

.PHONY: semgrep
semgrep: ## Run Semgrep security/quality scan (install via: make setup)
	@which semgrep > /dev/null 2>&1 || \
		(printf "$(YELLOW)semgrep not found — run: make setup$(RESET)\n" && exit 1)
	@semgrep scan --config=auto --lang go ./...

.PHONY: check
check: fmt vet lint ## Run all quality checks: fmt + vet + lint

.PHONY: security
security: semgrep ## Run Semgrep security scan

.PHONY: ci
ci: tidy fmt vet lint test ## Full CI pipeline locally (tidy → fmt → vet → lint → test)

# ─────────────────────────────────────────────────────────────────────────────
##@ Cross-Platform Builds
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: build-all
build-all: ## Cross-compile for all supported platforms
	@printf "$(GREEN)Cross-compiling $(BINARY) $(VERSION)…$(RESET)\n"
	@mkdir -p $(OUTPUT_DIR)
	@CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY)-darwin-arm64  .
	@CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY)-darwin-amd64  .
	@CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY)-linux-amd64   .
	@CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY)-linux-arm64   .
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(OUTPUT_DIR)/$(BINARY)-windows-amd64.exe .
	@printf "$(GREEN)Binaries:$(RESET)\n"
	@ls -lh $(OUTPUT_DIR)/$(BINARY)-*

.PHONY: snapshot
snapshot: ## Build a dev snapshot (no tag required)
	@$(MAKE) build VERSION=dev-$(COMMIT)

# ─────────────────────────────────────────────────────────────────────────────
##@ Installation
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: install
install: build ## Install binary to /usr/local/bin (requires sudo)
	@printf "$(CYAN)Installing to $(INSTALL_DIR)/$(BINARY)…$(RESET)\n"
	@sudo cp $(OUTPUT_DIR)/$(BINARY) $(INSTALL_DIR)/$(BINARY)
	@printf "$(GREEN)Done — run: $(BINARY) --version$(RESET)\n"

.PHONY: install-user
install-user: build ## Install binary to $(GOPATH)/bin (no sudo needed)
	@printf "$(CYAN)Installing to $(GOBIN)/$(BINARY)…$(RESET)\n"
	@cp $(OUTPUT_DIR)/$(BINARY) $(GOBIN)/$(BINARY)
	@printf "$(GREEN)Done — ensure $(GOBIN) is in your PATH$(RESET)\n"

.PHONY: uninstall
uninstall: ## Remove binary from /usr/local/bin
	@printf "$(YELLOW)Removing $(INSTALL_DIR)/$(BINARY)…$(RESET)\n"
	@sudo rm -f $(INSTALL_DIR)/$(BINARY)
	@printf "Done.\n"

# ─────────────────────────────────────────────────────────────────────────────
##@ Dev Setup
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: setup
setup: ## Install all dev tools: golangci-lint, semgrep, pre-commit
	@printf "$(CYAN)Installing golangci-lint…$(RESET)\n"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
		| sh -s -- -b $(GOBIN) v1.64.8
	@printf "$(CYAN)Installing semgrep…$(RESET)\n"
	@pip3 install --quiet semgrep 2>/dev/null || pip install --quiet semgrep
	@printf "$(CYAN)Installing pre-commit…$(RESET)\n"
	@pip3 install --quiet pre-commit 2>/dev/null || pip install --quiet pre-commit
	@printf "$(GREEN)All tools installed. Run 'make hooks' to activate git hooks.$(RESET)\n"

.PHONY: hooks
hooks: ## Install pre-commit git hooks
	@which pre-commit > /dev/null 2>&1 || \
		(printf "$(YELLOW)pre-commit not found — run: make setup$(RESET)\n" && exit 1)
	@pre-commit install
	@printf "$(GREEN)Git hooks active — they will run on every commit.$(RESET)\n"

.PHONY: hooks-run
hooks-run: ## Run pre-commit hooks against all files right now
	@pre-commit run --all-files

# ─────────────────────────────────────────────────────────────────────────────
##@ Maintenance
# ─────────────────────────────────────────────────────────────────────────────

.PHONY: clean
clean: ## Remove compiled binaries from dist/
	@printf "$(YELLOW)Cleaning dist/…$(RESET)\n"
	@rm -f $(OUTPUT_DIR)/$(BINARY) $(OUTPUT_DIR)/$(BINARY)-*
	@printf "Done.\n"

.PHONY: clean-all
clean-all: clean ## Remove dist/ entirely and coverage artifacts
	@rm -rf $(OUTPUT_DIR) coverage.out

.PHONY: version
version: ## Print the current version string
	@printf "$(VERSION)\n"

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(CYAN)MCPSync$(RESET) — available targets\n\n"} \
	     /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-18s$(RESET) %s\n", $$1, $$2 } \
	     /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@printf "\n"
