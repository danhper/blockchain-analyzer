all: build test

build:
	@go build -o . ./...

test:
	@go test ./...
