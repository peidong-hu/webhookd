package travis

import (
	"encoding/json"
	"github.com/jacksgt/travishook"
	"net/http"

	. "github.com/vision-it/webhookd/logging"
	. "github.com/vision-it/webhookd/model"
	"github.com/vision-it/webhookd/mq"
)

const defaultTravisConfigServer string = "api.travis-ci.org"

type TravisHandler struct {
	WebhookHandler
	route    string
	exchange string
}

func queueMessage(p travishook.WebhookPayload) string {
	var m MQMessage
	m.Version = MQMessageVersion
	m.Repository = p.Repository.Name
	m.Branch = p.Branch
	m.Commit = p.Commit
	m.Message = p.Message
	m.Author = p.AuthorName
	m.Trigger = "Travis Successful Build"

	message, _ := json.Marshal(&m)
	return string(message)
}

func New(route string, exchange string) (h *TravisHandler) {
	h = &TravisHandler{
		route:    route,
		exchange: exchange,
	}
	return h
}

/*
* Travis Webhook Delivery Format
* https://docs.travis-ci.com/user/notifications/#Webhooks-Delivery-Format
 */
func (h *TravisHandler) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {

	/* check request type */
	if reader.Method != "POST" {
		/* 405 Method Not Allowed */
		writer.Header().Set("Allow", "POST")
		http.Error(writer, http.StatusText(405), 405)
		Lg(1, "405: %s - %s\n", reader.Method, reader.URL)
		return
	}

	/* check content-type (header) */
	contentType := reader.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		/* 415 Unsupported Media Type */
		http.Error(writer, http.StatusText(415), 415)
		Lg(1, "415: %s - %s (Content-Type: %s)\n", reader.Method, reader.URL, contentType)
		return
	}

	/* get payload */
	rawPayload := reader.FormValue("payload")
	if rawPayload == "" {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "400: %s - %s (Empty Payload)\n", reader.Method, reader.URL)
		return
	}

	/* verify signature */
	signature := reader.Header.Get("Signature")
	if signature == "" {
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "400: %s - %s (Missing Signature)", reader.Method, reader.URL)
		return
	}

	err := travishook.CheckSignature(signature, []byte(rawPayload), defaultTravisConfigServer)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "Travis signature check failed: %s\n", err)
		return
	}

	/* json-decode payload */
	payload := travishook.WebhookPayload{}
	err = json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		Lg(0, "400: %s - %s (Error decoding JSON: %s)\n", reader.Method, reader.URL, err)
		return
	}

	/* check build status */
	if payload.Status == 0 && payload.StatusMessage == "Passed" {
		/* build successful */

		message := queueMessage(payload)

		err = mq.Publish(message, h.exchange)
		if err != nil {
			/* 500 Internal Server Error */
			http.Error(writer, http.StatusText(500), 500)
			Lg(0, "Publishing message %s failed: %s\n", message, err)
			return
		}

	} else {
		Lg(2, "Ignoring failed build %s from %s", payload.Number, payload.Repository.Name)
	}

	/* close HTTP stream */
	writer.WriteHeader(200)
	writer.Write([]byte("OK\n"))

	return
}
