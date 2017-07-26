package travis

import (
	"encoding/json"
	"net/http"

	"github.com/jacksgt/travishook"
	. "github.com/vision-it/webhookd/logging"
	. "github.com/vision-it/webhookd/model"
	"github.com/vision-it/webhookd/mq"
	"time"
)

const defaultTravisConfigServer string = "api.travis-ci.org"

type TravisHandler struct {
	WebhookHandler
	route    string
	exchange string
}

func queueMessage(p travisPayload) string {
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
	payload := travisPayload{}
	err = json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		Lg(0, "400: %s - %s (Error decoding JSON: %s)\n", reader.Method, reader.URL, err)

		/* for debugging purposes */
		print(rawPayload)

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

type travisPayload struct {
	ID     int    `json:"id"`
	Number string `json:"number"`
	Config struct {
		Sudo     bool     `json:"sudo"`
		Language string   `json:"language"`
		Cache    string   `json:"cache"`
		Script   []string `json:"script"`
		Matrix   struct {
			FastFinish bool `json:"fast_finish"`
			Include    []struct {
				Rvm         string `json:"rvm"`
				Env         string `json:"env"`
				BundlerArgs string `json:"bundler_args"`
				Dist        string `json:"dist"`
				Services    string `json:"services,omitempty"`
				Sudo        string `json:"sudo,omitempty"`
			} `json:"include"`
		} `json:"matrix"`
		Notifications struct {
			Email    bool `json:"email"`
			Webhooks struct {
				Urls      []string `json:"urls"`
				OnSuccess string   `json:"on_success"`
				OnFailure string   `json:"on_failure"`
				OnStart   string   `json:"on_start"`
				OnCancel  string   `json:"on_cancel"`
				OnError   string   `json:"on_error"`
			} `json:"webhooks"`
		} `json:"notifications"`
		Result string `json:".result"`
		Group  string `json:"group"`
		Dist   string `json:"dist"`
	} `json:"config"`
	Type              string      `json:"type"`
	State             string      `json:"state"`
	Status            int         `json:"status"`
	Result            int         `json:"result"`
	StatusMessage     string      `json:"status_message"`
	ResultMessage     string      `json:"result_message"`
	StartedAt         time.Time   `json:"started_at"`
	FinishedAt        time.Time   `json:"finished_at"`
	Duration          int         `json:"duration"`
	BuildURL          string      `json:"build_url"`
	CommitID          int         `json:"commit_id"`
	Commit            string      `json:"commit"`
	BaseCommit        string      `json:"base_commit"`
	HeadCommit        interface{} `json:"head_commit"`
	Branch            string      `json:"branch"`
	Message           string      `json:"message"`
	CompareURL        string      `json:"compare_url"`
	CommittedAt       time.Time   `json:"committed_at"`
	AuthorName        string      `json:"author_name"`
	AuthorEmail       string      `json:"author_email"`
	CommitterName     string      `json:"committer_name"`
	CommitterEmail    string      `json:"committer_email"`
	PullRequest       bool        `json:"pull_request"`
	PullRequestNumber interface{} `json:"pull_request_number"`
	PullRequestTitle  interface{} `json:"pull_request_title"`
	Tag               interface{} `json:"tag"`
	Repository        struct {
		ID        int         `json:"id"`
		Name      string      `json:"name"`
		OwnerName string      `json:"owner_name"`
		URL       interface{} `json:"url"`
	} `json:"repository"`
	Matrix []struct {
		ID           int    `json:"id"`
		RepositoryID int    `json:"repository_id"`
		ParentID     int    `json:"parent_id"`
		Number       string `json:"number"`
		State        string `json:"state"`
		Config       struct {
			Sudo          bool     `json:"sudo"`
			Language      string   `json:"language"`
			Cache         string   `json:"cache"`
			Script        []string `json:"script"`
			Notifications struct {
				Email    bool `json:"email"`
				Webhooks struct {
					Urls      []string `json:"urls"`
					OnSuccess string   `json:"on_success"`
					OnFailure string   `json:"on_failure"`
					OnStart   string   `json:"on_start"`
					OnCancel  string   `json:"on_cancel"`
					OnError   string   `json:"on_error"`
				} `json:"webhooks"`
			} `json:"notifications"`
			Result      string   `json:".result"`
			Group       string   `json:"group"`
			Dist        string   `json:"dist"`
			BundlerArgs string   `json:"bundler_args"`
			Env         []string `json:"env"`
			Rvm         string   `json:"rvm"`
			Os          string   `json:"os"`
		} `json:"config"`
		Status         int         `json:"status"`
		Result         int         `json:"result"`
		Commit         string      `json:"commit"`
		Branch         string      `json:"branch"`
		Message        string      `json:"message"`
		CompareURL     string      `json:"compare_url,omitempty"`
		StartedAt      interface{} `json:"started_at,omitempty"`
		FinishedAt     interface{} `json:"finished_at,omitempty"`
		CommittedAt    time.Time   `json:"committed_at,omitempty"`
		AuthorName     string      `json:"author_name,omitempty"`
		AuthorEmail    string      `json:"author_email,omitempty"`
		CommitterName  string      `json:"committer_name,omitempty"`
		CommitterEmail string      `json:"committer_email,omitempty"`
		AllowFailure   bool        `json:"allow_failure,omitempty"`
	} `json:"matrix"`
}
