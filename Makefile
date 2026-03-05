.PHONY: build test lint clean

build:
	go build -o bin/gtm-users ./cmd/gtm-users/

test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out
