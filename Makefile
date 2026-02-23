BINARY  := metamodel
GO      := go

.PHONY: all build test test-race fmt clean

all: build

## build: compile the binary
build:
	$(GO) build -o $(BINARY) .

## test: run tests without race detector
test:
	$(GO) test ./... -count=1 -coverprofile=coverage.out

## test-race: run tests with the race detector
test-race:
	$(GO) test ./... -race -count=1

## fmt: format all Go source files
fmt:
	$(GO) fmt ./...

## clean: remove build artifacts and cache
clean:
	$(GO) clean -cache
	rm -f $(BINARY) coverage.out coverage.html

