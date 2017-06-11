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


func setRoutes(routePrefix string, h HooksConfig) {

	/* GitHub */
	var defaultGithubRoute = h.Github[0].Route
	var defaultGithubSecret = h.Github[0].Secret
	var defaultGithubExchange = h.Github[0].Exchange

	if len(h.Github) > 1 {
		for i := range h.Github {
			var g GithubHandler
			e := h.Github[i]
			if e.Route == "" {
				g.Route = routePrefix + defaultGithubRoute
			} else {
				g.Route = routePrefix + e.Route
			}

			g.Secret = e.Secret
			if g.Secret == "" {
				g.Secret = defaultGithubSecret
			}

			g.Exchange = e.Exchange
			if g.Exchange == "" {
				g.Exchange = defaultGithubExchange
			}

			http.HandleFunc(g.Route, g.HandlerFunc)
		}
	}

	/* Gitlab */
	var defaultGitlabRoute = h.Gitlab[0].Route
	var defaultGitlabSecret = h.Gitlab[0].Secret
	var defaultGitlabExchange = h.Gitlab[0].Exchange

	for i := range h.Gitlab {
		var g GitlabHandler
		e := h.Gitlab[i]
		if e.Route == "" {
			g.Route = routePrefix + defaultGitlabRoute
		} else {
			g.Route = routePrefix + e.Route
		}

		g.Secret = e.Secret
		if g.Secret == "" {
			g.Secret = defaultGitlabSecret
		}

		g.Exchange = e.Exchange
		if g.Exchange == "" {
			g.Exchange = defaultGitlabExchange
		}

		log.Printf("Assigning route %s to Gitlab Handler", g.Route)
		http.HandleFunc(g.Route, g.HandlerFunc)
	}

}

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

	setRoutes(CONFIG.RoutePrefix, CONFIG.Hooks)

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
