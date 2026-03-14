.PHONY: build test test-cover lint vet clean

build:
	go build -o bin/gtm-users ./cmd/gtm-users/

test:
	go test ./... -v -race

test-cover:
	go test ./... -race -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

lint:
	golangci-lint run

vet:
	go vet ./...

clean:
	rm -rf bin/ coverage.out
