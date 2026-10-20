.PHONY: build test test-race bench lint vet embed-sync doctor gate clean release-check harness-probe harness-test harness-test-all harness-triage harness-clean harness-clean-sandboxes

VERSION ?= 0.1.0
LDFLAGS ?= -s -w -X main.version=$(VERSION)
export CGO_ENABLED ?= 0

ifeq ($(OS),Windows_NT)
  EXE := .exe
else
  EXE :=
endif

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o vck$(EXE) ./cmd/vivechak/
	go build -trimpath -ldflags "$(LDFLAGS)" -o vivechak$(EXE) ./cmd/vivechak/

doctor:
	go run ./cmd/vivechak doctor

gate:
	go run ./cmd/vivechak gate --track all

test:
	go test -count=1 ./...

test-race:
	CGO_ENABLED=1 go test -count=1 -race ./...

bench:
	go test -bench Benchmark -benchmem ./internal/core

vet:
	go vet ./...

lint:
	golangci-lint run ./...

embed-sync:
	go test -v -count=1 ./internal/embed -run TestEmbeddedFilesMatchRoot
	@echo "✓ Embed sync verified"

clean:
	@rm -f vivechak vck vivechak.exe vck.exe coverage* *.out
	@echo "✓ Clean complete"

release-check: vet lint test embed-sync
	@echo "✓ Release checks passed"

harness-probe:
	node scripts/run_harness.mjs --probe

harness-test:
	node scripts/run_harness.mjs --auto --batch core

harness-test-all:
	node scripts/run_harness.mjs --auto --batch all

harness-triage:
	node scripts/run_harness.mjs --triage

harness-clean:
	node scripts/run_harness.mjs --clean

harness-clean-sandboxes:
	node scripts/run_harness.mjs --clean-sandboxes
