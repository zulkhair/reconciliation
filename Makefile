.PHONY: build run test lint lint-install lint-fix 

# Default Go parameters
GOFLAGS := -v
TESTFLAGS := -v -race -cover

lint_version=v1.62.2

# Build binary
build:
	go build $(GOFLAGS) -o bin/reconciliation cmd/main.go

# Run the application
# Required parameters:
# - system: Path to system transaction CSV file
# - bank: Path to bank statement CSV files
# - start: Start date (YYYY-MM-DD)
# - end: End date (YYYY-MM-DD)
# Optional:
# - output: Path to output JSON file
run:
	@if [ -z "$(system)" ] || [ -z "$(bank)" ] || [ -z "$(start)" ] || [ -z "$(end)" ]; then \
		echo "Usage: make run system=<system-file> bank=<bank-file> start=<start-date> end=<end-date> [output=<output-file>]"; \
		exit 1; \
	fi
	go run $(GOFLAGS) cmd/main.go -s $(system) -b $(bank) -t $(start) -e $(end) $(if $(output),-o $(output))

# Run tests
test:
	go test $(TESTFLAGS) ./...

lint-install:
	@echo "--> Checking if golangci-lint $(lint_version) is installed"
	@installed_version=$$(golangci-lint --version 2> /dev/null | awk '{print $$4}') || true; \
	if [ "$$installed_version" != "$(lint_version)" ]; then \
		echo "--> Installing golangci-lint $(lint_version)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(lint_version); \
	else \
		echo "--> golangci-lint $(lint_version) is already installed"; \
	fi

lint:
	@$(MAKE) lint-install
	@echo "--> Running linter"
	@go list -f '{{.Dir}}/...' -m | xargs golangci-lint run  --timeout=10m --concurrency 8 -v

lint-fix:
	@$(MAKE) lint-install
	@echo "--> Running linter"
	@go list -f '{{.Dir}}/...' -m | xargs golangci-lint run  --timeout=10m --fix --concurrency 8 -v
