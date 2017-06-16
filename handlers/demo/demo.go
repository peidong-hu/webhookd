// +build ignore

package demo

import (
	"encoding/json"
	. "github.com/vision-it/webhookd/logging"
	. "github.com/vision-it/webhookd/model"
	"net/http"
)

type testPayload struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Author     string `json:"author"`
	Message    string `json:"message"`
}

func queueMessageFromTest(p testPayload) (m MQMessage) {
	m.Version = MQMessageVersion
	m.Repository = p.Repository
	m.Branch = p.Branch
	m.Commit = "cafebabe"
	m.Message = p.Message
	m.Author = p.Author
	m.Trigger = "Test-Webhook"

	return m
}

func testHandler(writer http.ResponseWriter, reader *http.Request) {
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
	var payload TestPayload
	err := json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		Lg(0, "400: %s - %s (Error decoding JSON: %s)\n", reader.Method, reader.URL, err.Error)
		return
	}

	/* generate queue message */
	rawMessage := queueMessageFromTest(payload)

	/* convert struct to JSON string */
	message, err := json.Marshal(rawMessage)
	if err != nil {
		http.Error(writer, http.StatusText(500), 500)
		Lg(0, "500: %s - %s (Error encoding JSON: %s)\n", reader.Method, reader.URL, err.Error)
		return
	}

	/* publish message to MQ */
	err = publishMessage(MQCHANNEL, string(message))
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
