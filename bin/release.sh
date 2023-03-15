#!/bin/bash

mkdir -p out

go generate ./...
go build -o out/fuddle cmd/fuddle/main.go
