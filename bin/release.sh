#!/bin/bash

VERSION=0.0.0

date=$(date '+%Y%m%d')
echo $VERSION-$date-$(git rev-parse --short HEAD) > pkg/build/REVISION

mkdir -p out

go generate ./...
go build -o out/fuddle main.go
