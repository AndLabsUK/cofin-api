all: bin/api bin/document_fetcher bin/market_fetcher

bin/api:
	@echo "Building API"
	@go build -o bin/api ./cmd/api

bin/document_fetcher:
	@echo "Building Document Fetcher"
	@go build -o bin/document_fetcher ./cmd/document_fetcher

bin/market_fetcher:
	@echo "Building Market Fetcher"
	@go build -o bin/market_fetcher ./cmd/market_fetcher