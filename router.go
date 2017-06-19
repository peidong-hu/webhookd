package main

import (
	. "github.com/vision-it/webhookd/config"
	"github.com/vision-it/webhookd/handlers/demo"
	"github.com/vision-it/webhookd/handlers/gitlab"
	"log"
	"net/http"
)

func setRoutes(routePrefix string, h *HooksConfig) (mux *http.ServeMux) {
	mux = http.NewServeMux()

	setGitlabRoutes(mux, routePrefix, h)
	setDemoRoutes(mux, routePrefix, h)

	return mux
}

func setGitlabRoutes(mux *http.ServeMux, routePrefix string, h *HooksConfig) {
	/* retrieve defaults from first field */
	var defaultRoute = h.Gitlab[0].Route
	var defaultSecret = h.Gitlab[0].Secret
	var defaultExchange = h.Gitlab[0].Exchange

	for _, v := range h.Gitlab {
		var r, s, e string
		if v.Route == "" {
			r = routePrefix + defaultRoute
		} else {
			r = routePrefix + v.Route
		}

		s = v.Secret
		if s == "" {
			s = defaultSecret
		}

		e = v.Exchange
		if e == "" {
			e = defaultExchange
		}

		g := gitlab.New(r, s, e)

		log.Printf("Route %s -> Gitlab Handler", r)
		mux.Handle(r, g)
	}

}

func setDemoRoutes(mux *http.ServeMux, routePrefix string, h *HooksConfig) {
	/* retrieve defaults from first field */
	var defaultRoute = h.Demo[0].Route
	var defaultSecret = h.Demo[0].Secret
	var defaultExchange = h.Demo[0].Exchange

	for _, v := range h.Demo {
		var r, s, e string
		if v.Route == "" {
			r = routePrefix + defaultRoute
		} else {
			r = routePrefix + v.Route
		}

		s = v.Secret
		if s == "" {
			s = defaultSecret
		}

		e = v.Exchange
		if e == "" {
			e = defaultExchange
		}

		g := demo.New(r, s, e)

		log.Printf("Route %s -> Demo Handler", r)
		mux.Handle(r, g)
	}

}