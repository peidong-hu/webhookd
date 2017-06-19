package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	. "github.com/vision-it/webhookd/config"
	. "github.com/vision-it/webhookd/logging"
	_ "github.com/vision-it/webhookd/model"
	"github.com/vision-it/webhookd/mq"
	"log"
	"net/http"
)

var CONFIG Config
var TESTHOOK bool
var MQCONNECTION *amqp.Connection
var MQCHANNEL *amqp.Channel

func main() {
	flag.IntVar(&VERBOSITY, "v", 1, "verbosity to use")
	flag.BoolVar(&TESTHOOK, "testhook", true, "enable test webhook at /webhooks/test")
	flag.Parse()

	CONFIG, err := LoadConfig("./webhookd.json")
	FailOnError(err, "Failed to load config: %s", err)

	err = ValidateConfig(CONFIG)
	FailOnError(err, "Failed to validate config: %s", err)

	/* connect to MQ */
	MQCONNECTION, MQCHANNEL = mq.Connect(CONFIG.MQ)
	defer MQCONNECTION.Close()
	defer MQCHANNEL.Close()

	mux := setRoutes(CONFIG.RoutePrefix, &CONFIG.Hooks)

	/* start HTTP server */
	listen := fmt.Sprintf("%s:%d", CONFIG.Address, CONFIG.Port)

	Lg(1, "Listening on %s\n", listen)

	log.Fatal(http.ListenAndServe(listen, mux))
}
