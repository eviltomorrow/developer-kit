#!/bin/bash

rm -rf package
mkdir -p package
mkdir -p package/certs

go build -o package/ca-create -ldflags "-s -w" ca-create/main.go
go build -o package/server -ldflags "-s -w" server/main.go
go build -o package/client -ldflags "-s -w" client/main.go

