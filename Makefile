# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=mcp-go-colly
BINARY_UNIX=bin/$(BINARY_NAME)_unix
BINARY_WINDOWS=bin/$(BINARY_NAME)_windows.exe
BINARY_DARWIN=bin/$(BINARY_NAME)_darwin
BINARY_PATH=bin/$(BINARY_NAME)

# Build flags
LDFLAGS=-ldflags "-s -w"

.PHONY: all build clean test run deps install

all: test build

build:
	mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_DARWIN)

test:
	$(GOTEST) -v ./...

run:
	mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) ./cmd/main.go
	./$(BINARY_PATH)

deps:
	$(GOGET) -v -t -d ./...

install:
	$(GOINSTALL) ./cmd/main.go

# Cross compilation
build-all:
	mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) ./cmd/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS) ./cmd/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DARWIN) ./cmd/main.go

# Development helpers
fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

lint:
	golangci-lint run

# Help command
help:
	@echo "Available commands:"
	@echo "  make build      - Build the binary"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make run        - Build and run the binary"
	@echo "  make deps       - Download dependencies"
	@echo "  make install    - Install the binary"
	@echo "  make build-all  - Build for all platforms"
	@echo "  make fmt        - Format code"
	@echo "  make vet        - Run go vet"
	@echo "  make lint       - Run linter"
	@echo "  make help       - Show this help message" 
