.PHONY: help test bench coverage lint fmt vet clean tools license check-license ci verify test-short race-test bench-compare coverage-func tidy profile-cpu profile-mem deps

# Default target
.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	$(GOTEST) -v -race ./...

test-short: ## Run tests with short flag
	$(GOTEST) -v -short ./...

race-test: ## Run comprehensive race detection tests
	./scripts/race-test.sh

bench: ## Run benchmarks
	$(GOTEST) -run='^$$' -bench=. -benchmem -benchtime=3s ./...

bench-compare: ## Run benchmarks and compare with baseline (requires benchstat)
	@if [ ! -f "old.txt" ]; then \
		echo "Running baseline benchmarks..."; \
		$(GOTEST) -run='^$$' -bench=. -benchmem -count=10 ./... > old.txt; \
	fi
	@echo "Running new benchmarks..."
	@$(GOTEST) -run='^$$' -bench=. -benchmem -count=10 ./... > new.txt
	@echo "Comparing benchmarks..."
	@benchstat old.txt new.txt

coverage: ## Generate coverage report
	$(GOTEST) -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-func: ## Show coverage by function
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -func=coverage.out

lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed, see https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=5m

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

tidy: ## Tidy go.mod
	$(GOMOD) tidy

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -f coverage.txt coverage.out coverage.html
	rm -f old.txt new.txt

ci: fmt vet lint test check-license ## Run all CI checks

tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.5.0
	go install github.com/google/addlicense@v1.1.1
	go install golang.org/x/perf/cmd/benchstat@latest

# Add license headers to all Go files
# Uses Google's addlicense tool (github.com/google/addlicense)
# Existing headers are preserved (no re-processing)
# Note: Requires addlicense in PATH or ~/go/bin - install with 'make tools'
license: ## Add MIT license headers to Go files (maintainers only)
	@echo "Adding license headers to Go files..."
	@if [ -z "$(HOME)" ] || [ ! -d "$(HOME)" ]; then \
		echo "Error: Invalid HOME directory (required for ~/go/bin tool installation)"; \
		echo "       If running in Docker/container, ensure HOME is set and ~/go/bin exists"; exit 1; \
	fi
	@PATH="$(HOME)/go/bin:$$PATH" && \
	command -v addlicense >/dev/null 2>&1 || { echo "Error: addlicense not found. Install with: make tools"; exit 1; } && \
	find . -name "*.go" -not -path "./vendor/*" -not -path "./examples/*" -print0 | \
		xargs -0 addlicense -c "Daniel Schmidt" -l mit -s=only -y 2025 -v
	@echo "License header addition complete!"

# Check that all Go files have license headers
# Uses addlicense in check mode - accepts ANY copyright holder name
# Only verifies that MIT + SPDX headers exist
# Note: Requires addlicense in PATH or ~/go/bin - install with 'make tools'
check-license: ## Check that all Go files have license headers
	@echo "Checking license headers..."
	@if [ -z "$(HOME)" ] || [ ! -d "$(HOME)" ]; then \
		echo "Error: Invalid HOME directory (required for ~/go/bin tool installation)"; \
		echo "       If running in Docker/container, ensure HOME is set and ~/go/bin exists"; exit 1; \
	fi
	@PATH="$(HOME)/go/bin:$$PATH" && \
	command -v addlicense >/dev/null 2>&1 || { echo "Error: addlicense not found. Install with: make tools"; exit 1; } && \
	find . -name "*.go" -not -path "./vendor/*" -not -path "./examples/*" -print0 | \
		if xargs -0 addlicense -check -l mit -s=only -y 2025; then \
			echo "✓ All Go files have license headers!"; \
		else \
			echo "✗ Some files are missing license headers. Run 'make license' to add them."; exit 1; \
		fi

profile-cpu: ## Run CPU profiling
	$(GOTEST) -run='^$$' -bench=. -benchmem -cpuprofile=cpu.prof ./...
	$(GOCMD) tool pprof -http=:8080 cpu.prof

profile-mem: ## Run memory profiling
	$(GOTEST) -run='^$$' -bench=. -benchmem -memprofile=mem.prof ./...
	$(GOCMD) tool pprof -http=:8080 mem.prof

deps: ## Download dependencies
	$(GOMOD) download

# Verify dependencies and license headers
verify: ## Verify dependencies and license headers
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "Checking license headers..."
	@$(MAKE) check-license
