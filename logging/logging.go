package logging

import (
	"log"
)

var VERBOSITY int = 1

func Lg(level int, format string, a ...interface{}) {
	if level <= VERBOSITY {
		log.Printf(format, a...)
	}
}

func FailOnError(err error, format string, a ...interface{}) {
	if err != nil {
		log.Fatalf(format, a...)
	}
}
