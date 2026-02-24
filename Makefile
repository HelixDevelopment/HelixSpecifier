.PHONY: build test test-race test-cover test-bench test-verbose fmt vet lint clean

build:
	go build ./...

test:
	GOMAXPROCS=2 go test -count=1 -p 1 ./...

test-race:
	GOMAXPROCS=2 go test -count=1 -race -p 1 ./...

test-cover:
	GOMAXPROCS=2 go test -count=1 -race -p 1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-bench:
	GOMAXPROCS=2 go test -count=1 -bench=. -benchmem ./...

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
