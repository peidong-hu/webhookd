package main

import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	. "github.com/vision-it/webhookd/config"
	"github.com/vision-it/webhookd/handlers/gitlab"
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

func setRoutes(routePrefix string, h HooksConfig) (mux *http.ServeMux) {
	mux = http.NewServeMux()

	/* GitHub */
	// var defaultGithubRoute = h.Github[0].Route
	// var defaultGithubSecret = h.Github[0].Secret
	// var defaultGithubExchange = h.Github[0].Exchange

	// if len(h.Github) > 1 {
	// 	for i := range h.Github {
	// 		var g GithubHandler
	// 		e := h.Github[i]
	// 		if e.Route == "" {
	// 			g.Route = routePrefix + defaultGithubRoute
	// 		} else {
	// 			g.Route = routePrefix + e.Route
	// 		}

	// 		g.Secret = e.Secret
	// 		if g.Secret == "" {
	// 			g.Secret = defaultGithubSecret
	// 		}

	// 		g.Exchange = e.Exchange
	// 		if g.Exchange == "" {
	// 			g.Exchange = defaultGithubExchange
	// 		}

	// 		http.HandleFunc(g.Route, g.HandlerFunc)
	// 	}
	// }

	/* Gitlab */
	var defaultGitlabRoute = h.Gitlab[0].Route
	var defaultGitlabSecret = h.Gitlab[0].Secret
	var defaultGitlabExchange = h.Gitlab[0].Exchange

	for _, v := range h.Gitlab {
		var r, s, e string
		if v.Route == "" {
			r = routePrefix + defaultGitlabRoute
		} else {
			r = routePrefix + v.Route
		}

		s = v.Secret
		if s == "" {
			s = defaultGitlabSecret
		}

		e = v.Exchange
		if e == "" {
			e = defaultGitlabExchange
		}

		g := gitlab.NewHandler(r, s, e)

		log.Printf("Assigning route %s to Gitlab Handler", r)
		mux.Handle(r, g)
	}

	return mux
}

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

	mux := setRoutes(CONFIG.RoutePrefix, CONFIG.Hooks)

	// if TESTHOOK {
	// 	/* register test webhook handler */
	// 	http.HandleFunc("/webhooks/test", testHandler)
	// }

	// /* register travis webhook handler */
	// http.HandleFunc("/webhooks/travis-ci", travisHandler)

	/* start HTTP server */
	listen := fmt.Sprintf("%s:%d", CONFIG.Address, CONFIG.Port)

	Lg(1, "Listening on %s\n", listen)

	log.Fatal(http.ListenAndServe(listen, mux))
}
