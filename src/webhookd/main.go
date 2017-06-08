package main

import (
	"fmt"
	"net/http"
	"log"
	"flag"
	"github.com/streadway/amqp"
)


var VERBOSITY int
var CONFIG Config
var TESTHOOK bool
var MQCONNECTION *amqp.Connection
var MQCHANNEL *amqp.Channel

func main() {
	flag.IntVar(&VERBOSITY, "v", 1, "verbosity to use")
	flag.BoolVar(&TESTHOOK, "testhook", true, "enable test webhook at /webhooks/test")
    flag.Parse()

	CONFIG, err := loadConfig("./webhookd.json")
	failOnError(err, "Failed to load config: %s", err)

	err = validateConfig(CONFIG)
	failOnError(err, "Failed to validate config: %s", err)

	/* connect to MQ */
	MQCONNECTION, MQCHANNEL = connectMQ(CONFIG.MQ)
	defer MQCONNECTION.Close()
	defer MQCHANNEL.Close()

	if TESTHOOK {
		/* register test webhook handler */
		http.HandleFunc("/webhooks/test", testHandler)
	}

	/* register travis webhook handler */
	http.HandleFunc("/webhooks/travis-ci", travisHandler)

	/* start HTTP server */
	listen := fmt.Sprintf("%s:%d", CONFIG.Address, CONFIG.Port)

	lg(1, "Listening on %s\n", listen)

	log.Fatal(http.ListenAndServe(listen, nil))
}