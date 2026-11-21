.PHONY: all build test clean run config

unexport GOROOT

BINARY_NAME=sortflix
BIN_DIR=bin
CONFIG_FILE=config.json
CONFIG_SAMPLE=config.example.json

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

config:
	@if [ ! -f $(CONFIG_FILE) ]; then \
		cp $(CONFIG_SAMPLE) $(CONFIG_FILE); \
		echo "✅ Created $(CONFIG_FILE) from $(CONFIG_SAMPLE)"; \
	else \
		echo "ℹ️  $(CONFIG_FILE) already exists. Keeping existing file."; \
	fi
