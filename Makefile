.PHONY: build test test-unit test-integration test-e2e test-security test-stress test-automation test-all test-race test-cover test-bench test-bench-full test-verbose fmt vet lint clean

build:
	go build ./...

test:
	GOMAXPROCS=2 go test -count=1 -p 1 ./...

test-unit:
	GOMAXPROCS=2 go test -count=1 -p 1 ./pkg/...

test-integration:
	GOMAXPROCS=2 go test -count=1 -p 1 ./tests/integration/...

test-e2e:
	GOMAXPROCS=2 go test -count=1 -p 1 ./tests/e2e/...

test-security:
	GOMAXPROCS=2 go test -count=1 -p 1 ./tests/security/...

test-stress:
	GOMAXPROCS=2 go test -count=1 -p 1 ./tests/stress/...

test-automation:
	GOMAXPROCS=2 go test -count=1 -p 1 ./tests/automation/...

test-all:
	GOMAXPROCS=2 go test -count=1 -p 1 ./...

test-race:
	GOMAXPROCS=2 go test -count=1 -race -p 1 ./...

test-cover:
	GOMAXPROCS=2 go test -count=1 -race -p 1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-bench:
	GOMAXPROCS=2 go test -count=1 -bench=. -benchmem ./...

test-bench-full:
	GOMAXPROCS=2 go test -count=1 -bench=. -benchmem -benchtime=1s ./...

test-verbose:
	GOMAXPROCS=2 go test -count=1 -race -p 1 -v ./...

fmt:
	gofmt -w .
	goimports -w .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

clean:
	rm -f coverage.out coverage.html
