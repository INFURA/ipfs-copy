install:
	export GO111MODULE=on
	env CGO_ENABLED=0 go install -ldflags="-s -w" ./cmd/ipfs-copy/

build:
	export GO111MODULE=on
	env CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ipfs-copy ./cmd/ipfs-copy/

install-linux:
	export GO111MODULE=on
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/ipfs-copy ./cmd/ipfs-copy/

run: install
	./bin/ipfs-copy

.PHONY: install install-linux