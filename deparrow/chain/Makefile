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
	@for proto_file in $(shell find ./proto -name '*.proto'); do \
		protoc \
			-I./proto \
			-I/tmp/protoc/include \
			-I$(shell go env GOPATH)/pkg/mod/github.com/cosmos/gogoproto@v1.7.0 \
			-I$(shell go env GOPATH)/pkg/mod/github.com/cosmos/cosmos-sdk@v0.50.12/proto \
			-I$(shell go env GOPATH)/pkg/mod/github.com/cosmos/cosmos-proto@v1.0.0-beta.5/proto \
			--go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$proto_file; \
	done
	@echo "Copying generated files to module types directories..."
	@cp -f dpc/proofofcompute*.pb.go x/proofofcompute/types/ 2>/dev/null || true
	@cp -f dpc/computemarket*.pb.go x/computemarket/types/ 2>/dev/null || true
	@cp -f dpc/agentwallet*.pb.go x/agentwallet/types/ 2>/dev/null || true
	@cp -f dpc/token*.pb.go x/token/types/ 2>/dev/null || true
	@rm -rf dpc/
	@echo "Proto generation complete!"

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
