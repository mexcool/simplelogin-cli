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

man:
	go run ./cmd/gen-man man/

clean:
	rm -rf bin/ man/

.PHONY: build install test man clean
