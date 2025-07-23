.PHONY: build lint security fmt test all

build:
	go build -o bin/ttoprep .

lint:
	golangci-lint run

security:
	gosec ./...

fmt:
	gofmt -s -w .

test:
	go test ./...

all: build lint security fmt test
