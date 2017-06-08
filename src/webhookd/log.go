package main

import (
	"log"
)

func lg(level int, format string, a ...interface{}) {
	if level <= VERBOSITY {
		log.Printf(format, a...)
	}
}

func failOnError(err error, format string, a ...interface{}) {
	if err != nil {
		log.Fatalf(format, a...)
	}
}
