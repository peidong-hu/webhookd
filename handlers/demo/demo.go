package demo

import (
	"encoding/json"
	. "github.com/vision-it/webhookd/logging"
	. "github.com/vision-it/webhookd/model"
	"github.com/vision-it/webhookd/mq"
	"net/http"
)

type DemoHandler struct {
	secret   string
	route    string
	exchange string
}

type testPayload struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Author     string `json:"author"`
	Message    string `json:"message"`
}

func queueMessageFromTest(p testPayload) string {
	var m MQMessage

	m.Version = MQMessageVersion
	m.Repository = p.Repository
	m.Branch = p.Branch
	m.Commit = "cafebabe"
	m.Message = p.Message
	m.Author = p.Author
	m.Trigger = "Test-Webhook"

	/* internal structure, no error message */
	raw, _ := json.Marshal(&m)

	return string(raw)
}

func New(route string, secret string, exchange string) (h *DemoHandler) {
	h = &DemoHandler{
		route:    route,
		secret:   secret,
		exchange: exchange,
	}

	return h
}

func (h *DemoHandler) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {
	Lg(2, "%s - %s [%s]: %s\n",
		reader.Method,
		reader.URL,
		reader.Header.Get("Content-Type"),
		reader.FormValue("payload"),
	)

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

	/* json-decode payload */
	var payload testPayload
	err := json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		Lg(0, "400: %s - %s (Error decoding JSON: %s)\n", reader.Method, reader.URL, err.Error)
		return
	}

	message := queueMessageFromTest(payload)

	/* publish message to MQ */
	err = mq.Publish(message, h.exchange)
	if err != nil {
		http.Error(writer, http.StatusText(500), 500)
		Lg(0, "500: %s - %s (Failed to publish message: %s)\n", reader.Method, reader.URL, err.Error)
		return
	}

	/* close HTTP stream */
	writer.WriteHeader(200)
	writer.Write([]byte("OK\n"))

	return

}
