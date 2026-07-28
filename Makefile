IMPORT_PATH = github.com/moeryomenko/tf-provider-cloud-hypervisor

# Version can be overridden
VERSION ?= dev
CACHE_DIR ?= .cache
TOOLS_DIR ?= tools
TOOLS_BIN_DIR ?= $(CACHE_DIR)/tools
CH_TEST_KERNEL ?= $(CACHE_DIR)/vmlinux
CH_TEST_INITRD ?= $(CACHE_DIR)/initrd.img
CH_KERNEL_URL ?= https://cloud-hypervisor.azureedge.net/pub/ci-master/head/vmlinux
CH_INITRD_URL ?= https://cloud-hypervisor.azureedge.net/pub/ci-master/head/initrd.img

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Default target
.DEFAULT_GOAL := help

.PHONY: help build test testacc lint fmt clean sweep testdeps-acc

help: ## Display this help message
	@echo "Terraform Provider Cloud-Hypervisor - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the provider binary
	@echo "Building terraform-provider-cloud-hypervisor..."
	@GOWORK=off go build $(LDFLAGS) -o terraform-provider-cloud-hypervisor

test: ## Run unit tests with coverage (serialized to avoid temp-file races in chproc tests)
	@echo "Running unit tests..."
	@GOWORK=off go test -p=1 -count=1 ./... -coverprofile=coverage.out

testacc: ## Run acceptance tests (requires running Cloud-Hypervisor)
	@echo "Running acceptance tests..."
	@TF_ACC=1 GOWORK=off go test -count=1 -timeout 30m ./internal/provider/

lint: $(TOOLS_BIN_DIR)/golangci-lint ## Run golangci-lint
	@echo "Running golangci-lint..."
	@GOWORK=off $(TOOLS_BIN_DIR)/golangci-lint run --fix ./...

fmt: ## Format code with gofmt (basic fallback for initial scaffolding)
	@echo "Formatting code..."
	@gofmt -w -s .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f terraform-provider-cloud-hypervisor
	@GOWORK=off go clean

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
	@GOWORK=off go test -sweep= ./internal/provider/ -timeout 10m

$(TOOLS_BIN_DIR)/golangci-lint: $(TOOLS_DIR)/go.mod
	@mkdir -p $(TOOLS_BIN_DIR)
	@cd $(TOOLS_DIR) && go build -o ../$(TOOLS_BIN_DIR)/golangci-lint github.com/golangci/golangci-lint/v2/cmd/golangci-lint
