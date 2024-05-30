#!/bin/bash

go build -o targets/https2http cmd/https2http/main.go
go build -o targets/httpdumpheader cmd/httpdumpheader/main.go


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o targets/https2http.linux.amd64 cmd/https2http/main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o targets/httpdumpheader.linux.amd64 cmd/httpdumpheader/main.go