package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"github.com/jacksgt/travishook"
)

const defaultTravisConfigServer string = "api.travis-ci.org"

/*
* Travis Webhook Delivery Format
* https://docs.travis-ci.com/user/notifications/#Webhooks-Delivery-Format
*/
func travisHandler(writer http.ResponseWriter, reader *http.Request) {
	lg(1, "%s - %s [%s]: %s\n",
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
		lg(1, "405: %s - %s\n", reader.Method, reader.URL)
		return
	}

	/* check content-type (header) */
	contentType := reader.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		/* 415 Unsupported Media Type */
		http.Error(writer, http.StatusText(415), 415)
		lg(1, "415: %s - %s (Content-Type: %s)\n", reader.Method, reader.URL, contentType)
		return
	}

	/* get payload */
	rawPayload := reader.FormValue("payload")
	if rawPayload == "" {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		lg(1, "400: %s - %s (Empty Payload)\n", reader.Method, reader.URL)
		return
	}

	lg(2, "%s", rawPayload)

	/* verify signature */
	signature := reader.Header.Get("Signature")
	if signature == "" {
		http.Error(writer, http.StatusText(400), 400)
		lg(1, "400: %s - %s (Missing Signature)", reader.Method, reader.URL)
		return
	}

	err := travishook.CheckSignature(signature, []byte(rawPayload), defaultTravisConfigServer)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		lg(1, "Travis signature check failed: %s\n", err)
		return
	}

	/* json-decode payload */
	payload := travishook.WebhookPayload{}
	err = json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		lg(0, "400: %s - %s (Error decoding JSON: %s)\n", reader.Method, reader.URL, err)
		return
	}

	lg(2, "%s: %s (Commit #%s on '%s' by '%s')\n",
		payload.Repository.Name,
		payload.StatusMessage,
		payload.Commit,
		payload.Branch,
		payload.CommitterName,
	)

	/* check build status */
	if payload.Status == 0 && payload.StatusMessage == "Passed" {
		/* build successful */
		fmt.Println("Build successful")

		//message := queueMessageFromTravis(payload)
		message := "{ \"version\": \"0.1\" }"

		err := publishMessage(MQCHANNEL, message)

		if err != nil {
			/* 500 Internal Server Error */
			http.Error(writer, http.StatusText(500), 500)
			lg(0, "Publishing message %s failed: %s\n", message, err)
			return
		}

	} else {
		lg(2, "Build not successful - ignoring")
	}

	/* close HTTP stream */
	writer.WriteHeader(200)
	writer.Write([]byte("OK\n"))

	return
}
