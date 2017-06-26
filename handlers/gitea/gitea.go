package gitea

import (
	"encoding/json"
	. "github.com/vision-it/webhookd/logging"
	. "github.com/vision-it/webhookd/model"
	"github.com/vision-it/webhookd/mq"
	"io/ioutil"
	"net/http"
	"strings"
)

type GiteaHandler struct {
	WebhookHandler
	route    string
	secret   string
	exchange string
}

func New(route string, secret string, exchange string) (h *GiteaHandler) {
	h = &GiteaHandler{
		route:    route,
		secret:   secret,
		exchange: exchange,
	}
	return h
}

func queueMessage(p GiteaPayload) string {
	var m MQMessage
	branchSlice := strings.Split(p.Ref, "/")
	branch := branchSlice[len(branchSlice)-1]

	m.Version = MQMessageVersion
	m.Repository = p.Repository.Name
	m.Branch = branch
	m.Commit = p.Commits[0].ID
	m.Message = p.Commits[0].Message
	m.Author = p.Commits[0].Author.Username
	m.Trigger = "Gitea Push"

	message, _ := json.Marshal(&m)
	return string(message)
}

func (h *GiteaHandler) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {

	/* check request type */
	if reader.Method != "POST" {
		/* 405 Method Not Allowed */
		writer.Header().Set("Allow", "POST")
		http.Error(writer, http.StatusText(405), 405)
		Lg(1, "Invalid request type: %s\n", reader.Method)
		return
	}

	/* check Gitea Event */
	if reader.Header.Get("X-Gitea-Event") != "push" {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "Invalid Gitea Event: %s\n", reader.Header.Get("X-Gitea-Event"))
		return
	}

	/* get payload (depending on content type) */
	var rawPayload string
	switch reader.Header.Get("Content-Type") {
	case "application/json":
		body, err := ioutil.ReadAll(reader.Body)
		if err != nil {
			http.Error(writer, http.StatusText(500), 500)
			Lg(1, "Error reading body: %s\n", err)
			return
		}
		rawPayload = string(body[:])
	case "application/x-www-form-urlencoded":
		rawPayload = reader.FormValue("payload")
	default:
		/* 415 Unsupported Media Type */
		http.Error(writer, http.StatusText(415), 415)
		Lg(1, "Invalid Content-Type: %s\n", reader.Header.Get("Content-Type"))
		return
	}

	if rawPayload == "" {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "Empty Payload\n")
		return
	}

	/* decode json payload */
	payload := GiteaPayload{}
	err := json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "Error decoding JSON: %s\n", err)
		return
	}

	/* check secret (if any) */
	if h.secret != "" && h.secret != payload.Secret {
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "Invalid secret: %s\n", payload.Secret)
		return
	}

	Lg(2, "Received Delivery '%s' (Event: %s) with Content-Type '%s'\n",
		reader.Header.Get("X-Gitea-Delivery"),
		reader.Header.Get("X-Gitea-Event"),
		reader.Header.Get("Content-Type"),
	)

	message := queueMessage(payload)

	err = mq.Publish(message, h.exchange)
	if err != nil {
		/* 500 Internal Server Error */
		http.Error(writer, http.StatusText(500), 500)
		Lg(0, "Publishing message %s failed: %s\n", message, err)
		return
	}

	writer.WriteHeader(200)
	writer.Write([]byte("OK"))

	return
}

type GiteaPayload struct {
	Secret     string `json:"secret"`
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	CompareURL string `json:"compare_url"`
	Commits    []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
		URL     string `json:"url"`
		Author  struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
	} `json:"commits"`
	Repository struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Website     string `json:"website"`
		Watchers    int    `json:"watchers"`
		Owner       struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Username string `json:"username"`
		} `json:"owner"`
		Private bool `json:"private"`
	} `json:"repository"`
	Pusher struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	} `json:"pusher"`
	Sender struct {
		Login     string `json:"login"`
		ID        int    `json:"id"`
		AvatarURL string `json:"avatar_url"`
	} `json:"sender"`
}
