#!/bin/bash

go generate ./...
go build ./...
go test ./...
golangci-lint run
