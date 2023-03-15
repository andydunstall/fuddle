#!/bin/bash

VERSION=0.0.0
echo $VERSION-$(git rev-parse --short HEAD) > pkg/build/REVISION

mkdir -p out

go generate ./...
go build -o out/fuddle cmd/fuddle/main.go
