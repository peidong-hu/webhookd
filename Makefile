#!/usr/bin/make

.PHONY = build build-dep clean

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
VERSION ?= $(shell git describe --always --dirty --tags)

BIN ?= webhookd

build: build-dep webhookd

build-dep:
	go get -d -t ./...

webhookd: $(SOURCES)
	go build -o $(BIN) -ldflags "-X main.VERSION=$(VERSION)"

listener: listen/*.go
	cd listen && go build -o ../listener

clean:
	go clean
	rm -f $(BIN) listener
