#!/usr/bin/make
.PHONY = build build-dep clean

BINARY ?= webhookd

# set GOPATH if not already set
GOPATH ?= $(shell pwd)
export GOPATH

build: build-dep webhookd

build-dep:
	go get -d -t ./...

webhookd:
	go build -o $(BINARY) $<

clean:
	rm -f $(BINARY)
