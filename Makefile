VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

build:
	go build -ldflags="$(LDFLAGS)" -o bin/sl ./cmd/sl

install: build
	cp bin/sl /usr/local/bin/sl

test:
	go test ./...

clean:
	rm -rf bin/

.PHONY: build install test clean
