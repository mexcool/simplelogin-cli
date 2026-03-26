VERSION ?= dev

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/sl ./cmd/sl

install: build
	cp bin/sl /usr/local/bin/sl

test:
	go test ./...

man:
	go run ./cmd/gen-man man/

clean:
	rm -rf bin/ man/

.PHONY: build install test man clean
