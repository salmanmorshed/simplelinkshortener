.PHONY: build

VERSION ?= devel
BUILD_FLAGS := -ldflags "-X github.com/salmanmorshed/simplelinkshortener/internal/cfg.Version=$(VERSION)"
OUTPUT_BIN := bin/simplelinkshortener

build:
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(OUTPUT_BIN)_linux_x64 ./cmd/simplelinkshortener
