@echo off
set GO111MODULE=on
set GOFLAGS=-mod=vendor
go mod vendor
:: strip debug info during build
go build -ldflags="-s -w" .
