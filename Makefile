.PHONY: build lint security fmt test all setup

build:
	go build -o bin/ttoprep .

lint:
	golangci-lint run

security:
	gosec ./...

fmt:
	gofmt -s -w .

test:
	go test ./...

all: build lint security fmt test

# Version variables
EXPECTED_GOLANGCI_LINT_VERSION := 2.1.6
EXPECTED_GOSEC_VERSION       := 2.22.7

.PHONY: setup
setup:
	@echo "Checking prerequisites..."

	@ACTUAL_GOLANGCI_LINT_VERSION=$$(golangci-lint --version 2>/dev/null | head -1 | awk '{print $$4}' || echo ""); \
	if [ "$$ACTUAL_GOLANGCI_LINT_VERSION" != "$(EXPECTED_GOLANGCI_LINT_VERSION)" ]; then \
	  echo >&2 "golangci-lint not found or version mismatch: expected $(EXPECTED_GOLANGCI_LINT_VERSION), got $$ACTUAL_GOLANGCI_LINT_VERSION. Installing..."; \
	  go install github.com/golangci/golangci-lint/cmd/golangci-lint@v$(EXPECTED_GOLANGCI_LINT_VERSION) || { echo >&2 "Failed to install golangci-lint"; exit 1; }; \
	fi; \
	echo "  ✓ golangci-lint $(EXPECTED_GOLANGCI_LINT_VERSION) is installed"

	@ACTUAL_GOSEC_VERSION=$$(gosec --version 2>/dev/null | grep '^Version:' | awk '{print $$2}' || echo ""); \
	if [ "$$ACTUAL_GOSEC_VERSION" != "$(EXPECTED_GOSEC_VERSION)" ]; then \
	  echo >&2 "gosec not found or version mismatch: expected $(EXPECTED_GOSEC_VERSION), got $$ACTUAL_GOSEC_VERSION. Installing..."; \
	  curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v$(EXPECTED_GOSEC_VERSION) || { echo >&2 "Failed to install gosec"; exit 1; }; \
	fi; \
	echo "  ✓ gosec $(EXPECTED_GOSEC_VERSION) is installed"

	@echo "All prerequisites satisfied."
