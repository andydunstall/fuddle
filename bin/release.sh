#!/bin/bash

mkdir -p out

pushd console/ui
	npm install
	npm run build
	tar -czvf console.tar.gz build
popd

mv console/ui/console.tar.gz out/console.tar.gz

go generate ./...
go build -o out/fuddle cmd/fuddle/main.go
