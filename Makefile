.PHONY: help test bench coverage lint fmt vet clean

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

ci: fmt vet lint test ## Run all CI checks

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOGET) golang.org/x/perf/cmd/benchstat
	@echo "Tools installed successfully"

profile-cpu: ## Run CPU profiling
	$(GOTEST) -run='^$$' -bench=. -benchmem -cpuprofile=cpu.prof ./...
	$(GOCMD) tool pprof -http=:8080 cpu.prof

profile-mem: ## Run memory profiling
	$(GOTEST) -run='^$$' -bench=. -benchmem -memprofile=mem.prof ./...
	$(GOCMD) tool pprof -http=:8080 mem.prof

deps: ## Download dependencies
	$(GOMOD) download

verify: ## Verify dependencies
	$(GOMOD) verify
