.PHONY: run build test lint fmt

run:
	go run ./cmd/server

build:
	CGO_ENABLED=1 go build -tags fts5 -o bin/server ./cmd/server

test:
	go test ./...

lint:
	go vet ./...

fmt:
	gofmt -w .
