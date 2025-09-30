ifeq ($(OS),Windows_NT)
	CLEAR_COMMAND = @cls
else
	CLEAR_COMMAND = @clear
endif


build:
	@go build -o ./bin/exec

run: build
	@$(CLEAR_COMMAND)
	./bin/exec