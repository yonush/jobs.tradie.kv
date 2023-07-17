@echo off
set GO111MODULE=on
set GOFLAGS=-mod=mod
set GOWORK=off
go mod download
:: strip debug info during build
go build -ldflags="-s -w" .
