@echo off
set GO111MODULE=on
set GOFLAGS=-mod=vendor
set GOWORK=off
go mod vendor
:: strip debug info during build
go build -ldflags="-s -w" .
