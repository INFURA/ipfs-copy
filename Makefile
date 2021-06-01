#!/bin/bash
SHELL:=/usr/bin/env bash

build:
	export GO111MODULE=on
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ipfs-copy ./cmd/ipfs-copy/

install:
	export GO111MODULE=on
	env CGO_ENABLED=0 go install -ldflags="-s -w" ./cmd/ipfs-copy/

install-linux:
	export GO111MODULE=on
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/ipfs-copy ./cmd/ipfs-copy/

run: build
	source .env && ./bin/ipfs-copy

.PHONY: build install install-linux