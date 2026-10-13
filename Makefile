.PHONY: build test test-race bench lint vet embed-sync release-check

build:
	go build ./cmd/vivechak/

test:
	go test -count=1 ./...

test-race:
	go test -count=1 -race ./...

bench:
	go test -bench Benchmark -benchmem ./internal/core

vet:
	go vet ./...

lint:
	golangci-lint run ./...

embed-sync:
	go test -v -count=1 ./internal/embed -run TestEmbeddedFilesMatchRoot
	@echo "✓ Embed sync verified"

release-check: vet lint test embed-sync
	@echo "✓ Release checks passed"
