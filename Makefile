IMPORT_PATH = github.com/moeryomenko/tf-provider-cloud-hypervisor

# Version can be overridden
VERSION ?= dev
CACHE_DIR ?= .cache
CH_TEST_KERNEL ?= $(CACHE_DIR)/vmlinux
CH_TEST_INITRD ?= $(CACHE_DIR)/initrd.img
CH_KERNEL_URL ?= https://cloud-hypervisor.azureedge.net/pub/ci-master/head/vmlinux
CH_INITRD_URL ?= https://cloud-hypervisor.azureedge.net/pub/ci-master/head/initrd.img

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Default target
.DEFAULT_GOAL := help

.PHONY: help build test testacc lint fmt tidy vet clean sweep testdeps-acc check

help: ## Display this help message
	@echo "Terraform Provider Cloud-Hypervisor - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the provider binary
	@echo "Building tf-provider-cloud-hypervisor..."
	@go build -trimpath $(LDFLAGS) -o tf-provider-cloud-hypervisor .

test: ## Run all tests
	@go tool gotestsum --format-hide-empty-pkg -f testname -- -p=1 -vet=off -count=1 -timeout=1200s -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | grep ^total

testacc: ## Run acceptance tests (requires running Cloud-Hypervisor)
	@echo "Running acceptance tests..."
	@TF_ACC=1 go test -count=1 -timeout 30m ./internal/provider/

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@go tool golangci-lint run --fix ./...

fmt: ## Format Go source files
	@gofmt -s -w .
	@git status --short | grep '[A|M]' | grep -E -o "[^ ]*$$" | grep '\.go$$' | xargs -I{} go tool golines --base-formatter=gofumpt --ignore-generated --tab-len=1 --max-len=120 -w {}
	@git status --short | grep '[A|M]' | grep -E -o "[^ ]*$$" | grep '\.go$$' | xargs -I{} go tool goimports -local $(IMPORT_PATH) -w {}

tidy: ## Tidy go module dependencies
	@go mod tidy -v

vet: ## Run go vet
	@go vet ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f tf-provider-cloud-hypervisor coverage.out
	@go clean

testdeps-acc: ## Download acceptance test kernel+initrd fixtures to .cache/
	@echo "Preparing acceptance test dependencies..."
	@mkdir -p $(CACHE_DIR)
	@if [ -f "$(CH_TEST_KERNEL)" ]; then \
		echo "Using cached kernel: $(CH_TEST_KERNEL)"; \
	elif command -v curl >/dev/null 2>&1; then \
		echo "Downloading kernel from $(CH_KERNEL_URL) -> $(CH_TEST_KERNEL)"; \
		curl -fL --retry 3 --retry-delay 2 -o "$(CH_TEST_KERNEL)" "$(CH_KERNEL_URL)" || \
			echo "Warning: kernel download failed (network may be unavailable). Set CH_TEST_KERNEL env to a local path."; \
	else \
		echo "Warning: curl not available. Set CH_TEST_KERNEL env to a local path."; \
	fi
	@if [ -f "$(CH_TEST_INITRD)" ]; then \
		echo "Using cached initrd: $(CH_TEST_INITRD)"; \
	elif command -v curl >/dev/null 2>&1; then \
		echo "Downloading initrd from $(CH_INITRD_URL) -> $(CH_TEST_INITRD)"; \
		curl -fL --retry 3 --retry-delay 2 -o "$(CH_TEST_INITRD)" "$(CH_INITRD_URL)" || \
			echo "Warning: initrd download failed (network may be unavailable). Set CH_TEST_INITRD env to a local path."; \
	else \
		echo "Warning: curl not available. Set CH_TEST_INITRD env to a local path."; \
	fi
	@echo "Set CH_TEST_KERNEL=$(CH_TEST_KERNEL) and CH_TEST_INITRD=$(CH_TEST_INITRD) for acceptance tests."

sweep: ## Clean up leaked test resources (requires resource.TestMain in test files)
	@echo "Running sweepers..."
	@TF_ACC= go test -sweep= ./internal/provider/ -timeout 10m

check: lint vet test ## Run lint, vet, and tests (CI gate)
