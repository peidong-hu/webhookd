#!/usr/bin/make

.PHONY = build build-dep clean

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BIN ?= webhookd

build: build-dep webhookd

build-dep:
	go get -d -t ./...

webhookd: $(SOURCES)
	go build -o $(BIN)

listener: listen/*.go
	cd listen && go build -o ../listener

clean:
	go clean
	rm -f $(BIN) listener
