VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

build:
	go build -ldflags="$(LDFLAGS)" -o bin/sl ./cmd/sl
	cp bin/sl bin/simplelogin

PREFIX ?= /usr/local
install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 bin/sl $(DESTDIR)$(PREFIX)/bin/sl
	install -m 755 bin/simplelogin $(DESTDIR)$(PREFIX)/bin/simplelogin

test:
	go test ./...

man:
	go run ./cmd/gen-man man/

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -rf bin/ man/

.PHONY: build install test man lint fmt vet coverage clean
