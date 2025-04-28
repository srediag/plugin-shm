.PHONY: all build test lint vet tidy clean coverage mod-download help check build-examples run-examples coverage-examples ci

GO ?= go
GOLANGCI_LINT ?= golangci-lint

SRC_DIR=./plugin
BIN_DIR=bin
EXAMPLE_DIRS := $(shell find ./examples -type f -name 'main.go' -exec dirname {} \; | sort -u)

# The 'all' target does everything: lint, vet, test, coverage, build, build-examples, run-examples
all: check build build-examples run-examples

check: vet lint test coverage

build:
	$(GO) build $(SRC_DIR)/...

test:
	$(GO) test -v -race -cover $(SRC_DIR)/...

lint:
	$(GOLANGCI_LINT) run $(SRC_DIR)

vet:
	$(GO) vet $(SRC_DIR)/...

tidy:
	$(GO) mod tidy

clean:
	rm -rf coverage.out coverage.txt $(BIN_DIR)

coverage:
	$(GO) test -coverprofile=coverage.out $(SRC_DIR)/...
	$(GO) tool cover -func=coverage.out

mod-download:
	$(GO) mod download

# EXAMPLES TARGETS
build-examples:
	@mkdir -p $(BIN_DIR)
	@if [ -z "$(EXAMPLE_DIRS)" ]; then \
	  echo "No example applications to build."; \
	else \
	  for d in $(EXAMPLE_DIRS); do \
	    app=$$(basename $$d); \
	    parent=$$(basename $$(dirname $$d)); \
	    out=$(BIN_DIR)/$${parent}_$${app}; \
	    echo "Building $$d -> $$out"; \
	    $(GO) build -o $$out $$d; \
	  done; \
	fi

run-examples:
	@if [ -d $(BIN_DIR) ] && ls $(BIN_DIR)/* 1> /dev/null 2>&1; then \
	  for bin in $(BIN_DIR)/*; do \
	    echo "Running $$bin"; \
	    $$bin || echo "$$bin exited with code $$?"; \
	  done; \
	else \
	  echo "No example binaries to run."; \
	fi

coverage-examples:
	@if [ -z "$(EXAMPLE_DIRS)" ]; then \
	  echo "No example applications for coverage."; \
	else \
	  for d in $(EXAMPLE_DIRS); do \
	    echo "Testing coverage for $$d"; \
	    $(GO) test -coverprofile=coverage_$$app.out $$d; \
	    $(GO) tool cover -func=coverage_$$app.out; \
	  done; \
	fi

ci: all

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  all                Run full workflow: vet, lint, test, coverage, build, build-examples, run-examples (default)"
	@echo "  check              Run vet, lint, test, and coverage (plugin only)"
	@echo "  build              Build all Go code in ./plugin"
	@echo "  build-examples     Build all example applications to ./bin"
	@echo "  run-examples       Run all example binaries in ./bin (for experiments)"
	@echo "  coverage           Generate and show test coverage report for ./plugin"
	@echo "  coverage-examples  Generate and show test coverage for each example"
	@echo "  test               Run all tests with race and coverage (plugin only)"
	@echo "  lint               Run golangci-lint on ./plugin"
	@echo "  vet                Run go vet on ./plugin"
	@echo "  tidy               Run go mod tidy"
	@echo "  clean              Remove coverage and build artifacts"
	@echo "  mod-download       Download Go module dependencies"
	@echo "  ci                 Run the full workflow (alias for 'all')"
	@echo "  help               Show this help message"
