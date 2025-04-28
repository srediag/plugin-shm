.PHONY: all build test lint vet tidy clean coverage mod-download help check build-examples run-examples ci

GO ?= go
GOLANGCI_LINT ?= golangci-lint

# Dynamically set SRC_DIRS to avoid errors if ./examples is empty
SRC_DIRS := ./plugin
ifneq (,$(wildcard ./examples/*.go ./examples/*/*.go ./examples/*/*/*.go))
SRC_DIRS += ./examples
endif
BIN_DIR=bin

# The 'all' target does everything: lint, vet, test, coverage, build, build-examples, run-examples
all: check build build-examples run-examples

check: vet lint test coverage

build:
	$(GO) build ./...

build-examples:
	@mkdir -p $(BIN_DIR)
	@if find ./examples -type f -name main.go | grep -q .; then \
	  for d in $$(find ./examples -type f -name main.go | xargs -n1 dirname); do \
	    app=$$(basename $$d); \
	    parent=$$(basename $$(dirname $$d)); \
	    out=$(BIN_DIR)/$${parent}_$${app}; \
	    echo "Building $$d -> $$out"; \
	    $(GO) build -o $$out $$d; \
	  done; \
	else \
	  echo "No example applications to build."; \
	fi

run-experiments: run-examples

run-examples:
	@if [ -d $(BIN_DIR) ] && ls $(BIN_DIR)/* 1> /dev/null 2>&1; then \
	  for bin in $(BIN_DIR)/*; do \
	    echo "Running $$bin"; \
	    $$bin || echo "$$bin exited with code $$?"; \
	  done; \
	else \
	  echo "No example binaries to run."; \
	fi

test:
	$(GO) test -v -race -cover $(SRC_DIRS)

lint:
	$(GOLANGCI_LINT) run $(SRC_DIRS)

vet:
	$(GO) vet $(SRC_DIRS)

tidy:
	$(GO) mod tidy

clean:
	rm -rf coverage.out coverage.txt $(BIN_DIR)

coverage:
	$(GO) test -coverprofile=coverage.out $(SRC_DIRS)
	$(GO) tool cover -func=coverage.out

mod-download:
	$(GO) mod download

ci: all

help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  all            Run full workflow: vet, lint, test, coverage, build, build-examples, run-examples (default)"
	@echo "  check          Run vet, lint, test, and coverage"
	@echo "  build          Build all Go code"
	@echo "  build-examples Build all example applications to ./bin"
	@echo "  run-examples   Run all example binaries in ./bin (for experiments)"
	@echo "  test           Run all tests with race and coverage"
	@echo "  lint           Run golangci-lint on all code"
	@echo "  vet            Run go vet on all code"
	@echo "  tidy           Run go mod tidy"
	@echo "  clean          Remove coverage and build artifacts"
	@echo "  coverage       Generate and show test coverage report"
	@echo "  mod-download   Download Go module dependencies"
	@echo "  ci             Run the full workflow (alias for 'all')"
	@echo "  help           Show this help message"
