SHELL := /bin/bash

ifeq ($(GOPRIVATE),)
    export GOPRIVATE = 'github.com/qbox/pili,github.com/qbox/net-deftones'
endif

all: build

build:
	go build -v ./...

test:
	go test -v -failfast -parallel $(shell nproc) -timeout=2s -race ./...
