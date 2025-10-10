ifeq ($(OS),Windows_NT)
	CLEAR_COMMAND = @cls
else
	CLEAR_COMMAND = @clear
endif

.PHONY: all build run test test-verbose test-race test-race-verbose build-discovery run-discovery clean help

all: build

# ==============================================================================
# Go Application Commands
# ==============================================================================

test:
	@echo "=> Running Go tests..."
	@go test -run='^Test' ./...

test-verbose:
	@echo "=> Running Go tests (verbose)..."
	@go test -v -run='^Test' ./...

test-race:
	@echo "=> Running Go tests with race detector..."
	@go test -race -run='^Test' ./...

build:
	@echo "=> Building Go application..."
	@go build -o ./bin/exec .

run: build
	@$(CLEAR_COMMAND)
	./bin/exec

# ==============================================================================
# Rust Discovery Server Commands
# ==============================================================================

build-discovery:
	@echo "=> Building the Rust discovery server..."
	@cargo build --manifest-path ./discovery-server/Cargo.toml

run-discovery:
	@echo "=> Running the Rust discovery server..."
	@cargo run --release --manifest-path ./discovery-server/Cargo.toml

# ==============================================================================
# Cleanup & Help
# ==============================================================================

clean:
	@echo "=> Cleaning up..."
	@rm -f ./bin/exec
	@cargo clean --manifest-path ./discovery-server/Cargo.toml

help:
	@echo "Available commands:"
	@echo "  build             - Build the Go application"
	@echo "  run               - Run the Go application"
	@echo "  test              - Run Go unit tests"
	@echo "  test-verbose      - Run Go unit tests verbosely"
	@echo "  test-race         - Run Go unit tests with the race detector"
	@echo "  build-discovery   - Build the Rust discovery server"
	@echo "  run-discovery     - Run the Rust discovery server"
	@echo "  clean             - Clean up all build artifacts"
	@echo "  help              - Show this help message"