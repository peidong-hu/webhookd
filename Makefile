#!/usr/bin/make
.PHONY = build build-dep clean

BINARY ?= webhookd

# set GOPATH
GOPATH := $(shell pwd)
export GOPATH

build: build-dep webhookd

build-dep:
	go get -d -t ./...

webhookd: src/webhookd/*
	go build -o $(BINARY) $@

listen: src/listen/*
	go build $@

clean:
	rm -f $(BINARY) listen
