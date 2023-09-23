all: build

build:
	go build -v ./...

test:
	go test -v -failfast -parallel $(shell nproc) -timeout=2s -race ./...
