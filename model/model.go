package model

import (
	"net/http"
)

const MQMessageVersion string = "0.0"

type WebhookHandler interface {
	Handler(http.ResponseWriter, *http.Request)
	Route() string
}

type MQMessage struct {
	Version    string `json:"version"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Commit     string `json:"commit"`
	Message    string `json:"message"`
	Author     string `json:"author"`
	Trigger    string `json:"trigger"`
}
