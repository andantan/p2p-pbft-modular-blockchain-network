ifeq ($(OS),Windows_NT)
	CLEAR_COMMAND = @cls
else
	CLEAR_COMMAND = @clear
endif

test:
	@go test -run='^Test' ./...

test-verbose:
	@go test -v -run='^Test' ./...

test-race:
	@go test -race -run='^Test' ./...

test-race-verbose:
	@go test -v -race -run='^Test' ./...

build:
	@go build -o ./bin/exec

run: build
	@$(CLEAR_COMMAND)
	./bin/exec