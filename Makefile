.PHONY: all build test clean run

BINARY_NAME=sortflix
BIN_DIR=bin

all: test build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/sortflix

test:
	go test -v ./...

clean:
	rm -rf $(BIN_DIR)

run:
	go run ./cmd/sortflix

