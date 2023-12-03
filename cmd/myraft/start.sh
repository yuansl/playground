#!/bin/bash

if [ ! -x "./myraft" ]; then
    go build -ldflags=all=-w -v .
    docker build -t myraft:latest .
fi
