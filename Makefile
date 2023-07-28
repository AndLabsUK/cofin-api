all: bin/api bin/document_fetcher bin/market_fetcher

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: bin/api
bin/api:
	@echo "Building API"
	@go build -o bin/api ./cmd/api

.PHONY: bin/document_fetcher
bin/document_fetcher:
	@echo "Building Document Fetcher"
	@go build -o bin/document_fetcher ./cmd/document_fetcher

.PHONY: bin/market_fetcher
bin/market_fetcher:
	@echo "Building Market Fetcher"
	@go build -o bin/market_fetcher ./cmd/market_fetcher