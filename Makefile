all: deps build test

deps:
	@go get ./...

build:
	@go build -o . ./...

test:
	@go test ./...
