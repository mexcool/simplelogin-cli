VERSION ?= dev

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/sl ./cmd/sl

install: build
	cp bin/sl /usr/local/bin/sl

test:
	go test ./...

clean:
	rm -rf bin/

.PHONY: build install test clean
