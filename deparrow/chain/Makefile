#!/usr/bin/make -f

# DPC Chain Makefile
# Build and test commands for DEparrow Coin blockchain

BINARY_NAME := dpcd
VERSION := v1.0.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DIR := build
GO := go
GOFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT)"

# Default target
all: build

## build: Build the DPC daemon binary
build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(GOFLAGS) ./cmd/dpcd

## install: Install the binary to $GOPATH/bin
install:
	$(GO) install $(GOFLAGS) ./cmd/dpcd

## test: Run unit tests
test:
	$(GO) test -v ./...

## test-cover: Run tests with coverage
test-cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format Go code
fmt:
	$(GO) fmt ./...

## vet: Run go vet
vet:
	$(GO) vet ./...

## mod-tidy: Tidy go modules
mod-tidy:
	$(GO) mod tidy

## mod-download: Download dependencies
mod-download:
	$(GO) mod download

## deps: Install all dependencies
deps: mod-download

## clean: Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

## proto-gen: Generate protobuf code (requires protoc)
proto-gen:
	@echo "Generating protobuf files..."
	@for dir in $(shell find ./proto -name '*.proto' -exec dirname {} \; | sort -u); do \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			-I./proto $$dir/*.proto; \
	done

## init: Initialize a new node (for development)
init:
	./$(BUILD_DIR)/$(BINARY_NAME) init test-node --chain-id dpc-testnet-1

## keys: Generate a new key (for development)
keys:
	./$(BUILD_DIR)/$(BINARY_NAME) keys add validator --keyring-backend test

## start: Start the node (for development)
start:
	./$(BUILD_DIR)/$(BINARY_NAME) start

## help: Show this help
help:
	@echo "DPC Chain - DEparrow Coin Blockchain"
	@echo ""
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

.PHONY: all build install test test-cover lint fmt vet mod-tidy mod-download deps clean proto-gen init keys start help
