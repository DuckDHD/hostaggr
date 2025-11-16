.PHONY: run test build clean docker-build docker-up docker-down docker-logs help

# Default target
.DEFAULT_GOAL := help

# Binary output directory
BIN_DIR := bin
BINARY_NAME := hostaggr

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GORUN := $(GOCMD) run
GOTEST := $(GOCMD) test
GOCLEAN := $(GOCMD) clean

## run: Run the hotel aggregator server
run:
	$(GORUN) cmd/server/main.go

## test: Run all tests with race detection and verbose output
test:
	$(GOTEST) -v -race ./...

## build: Build the binary
build:
	mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) cmd/server/main.go

## clean: Remove build artifacts
clean:
	rm -rf $(BIN_DIR)
	$(GOCLEAN)

## docker-build: Build the Docker image
docker-build:
	docker-compose build

## docker-up: Start the Docker container
docker-up:
	docker-compose up

## docker-down: Stop and remove Docker containers
docker-down:
	docker-compose down

## docker-logs: View Docker container logs
docker-logs:
	docker-compose logs -f

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
