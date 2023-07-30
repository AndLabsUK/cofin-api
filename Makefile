all: build_macos build_linux
build_macos: bin/api-macos bin/document_fetcher-macos bin/market_fetcher-macos
build_linux: bin/api-linux bin/document_fetcher-linux bin/market_fetcher-linux

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: bin/api-macos
bin/api-macos:
	@echo "Building API for MacOS"
	@GOOS=darwin GOARCH=amd64 go build -o bin/api-amd64-darwin ./cmd/api

.PHONY: bin/document_fetcher-macos
bin/document_fetcher-macos:
	@echo "Building document fetcher for MacOS"
	@GOOS=darwin GOARCH=amd64 go build -o bin/document_fetcher-amd64-darwin ./cmd/document_fetcher

.PHONY: bin/market_fetcher-macos
bin/market_fetcher-macos:
	@echo "Building market fetcher for MacOS"
	@GOOS=darwin GOARCH=amd64 go build -o bin/market_fetcher-amd64-darwin ./cmd/market_fetcher

.PHONY: bin/api-linux
bin/api-linux:
	@echo "Building API for Linux"
	@GOOS=linux GOARCH=amd64 go build -o bin/api-amd64-linux ./cmd/api

.PHONY: bin/document_fetcher-linux
bin/document_fetcher-linux:
	@echo "Building document fetcher for Linux"
	@GOOS=linux GOARCH=amd64 go build -o bin/document_fetcher-amd64-linux ./cmd/document_fetcher

.PHONY: bin/market_fetcher-linux
bin/market_fetcher-linux:
	@echo "Building market fetcher for Linux"
	@GOOS=linux GOARCH=amd64 go build -o bin/market_fetcher-amd64-linux ./cmd/market_fetcher