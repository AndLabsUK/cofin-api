build:
	go build -o bin/ ./cmd/...

run: build
	./bin/$(APP_NAME)