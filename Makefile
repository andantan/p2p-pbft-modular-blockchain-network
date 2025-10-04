ifeq ($(OS),Windows_NT)
	CLEAR_COMMAND = @cls
else
	CLEAR_COMMAND = @clear
endif

test:
	@go test ./...

test-verbose:
	@go test -v ./...

test-race:
	@go test -race ./...

test-race-verbose:
	@go test -v -race ./...

build:
	@go build -o ./bin/exec

run: build
	@$(CLEAR_COMMAND)
	./bin/exec