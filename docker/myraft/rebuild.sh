#!/bin/bash
docker compose down -v

cp ~/go/bin/myraft  .
docker build -t myraft .

docker compose up -d
