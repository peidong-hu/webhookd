package gitlab

import (
	"encoding/json"
	. "github.com/vision-it/webhookd/logging"
	. "github.com/vision-it/webhookd/model"
	"github.com/vision-it/webhookd/mq"
	"net/http"
	"strings"
	"time"
)

/* Gitlab Webhooks: https://docs.gitlab.com/ce/user/project/integrations/webhooks.html */

type GitlabHandler struct {
	WebhookHandler
	secret   string
	route    string
	exchange string
}

func queueMessageFromPayload(p GitlabPayload) string {
	var m MQMessage
	branchSlice := strings.Split(p.Ref, "/")
	branch := branchSlice[len(branchSlice)-1]

	m.Version = MQMessageVersion
	m.Repository = p.Project.PathWithNamespace
	m.Branch = branch
	m.Commit = p.Commits[0].ID
	m.Message = p.Commits[0].Message
	m.Author = p.UserUsername
	m.Trigger = "Gitlab Push"

	/* internal structure, no error message */
	raw, _ := json.Marshal(&m)

	return string(raw)
}

/* generates a new Gitlab Handler */
func New(route string, secret string, exchange string) (h *GitlabHandler) {
	h = &GitlabHandler{
		route:    route,
		secret:   secret,
		exchange: exchange,
	}
	return h
}

func (h *GitlabHandler) ServeHTTP(writer http.ResponseWriter, reader *http.Request) {

	/* check request type */
	if reader.Method != "POST" {
		/* 405 Method Not Allowed */
		writer.Header().Set("Allow", "POST")
		http.Error(writer, http.StatusText(405), 405)
		Lg(1, "405: %s - %s\n", reader.Method, reader.URL)
		return
	}

	/* check Gitlab headers */
	event := reader.Header.Get("X-Gitlab-Event")
	if event == "" {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "400: %s - %s (Missing Gitlab Header)\n", reader.Method, reader.URL)
		return
	}

	/* only process push events */
	if event != "Push Hook" {
		writer.WriteHeader(200)
		writer.Write([]byte("OK\n"))
		Lg(1, "Ignoring Event %s for %s", event, reader.URL)
		return
	}

	/* verify secret */
	secret := reader.Header.Get("X-Gitlab-Token")
	if h.secret != "" && secret != h.secret {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "400: %s - %s (Invalid or missing secret)\n", reader.Method, reader.URL)
		return
	}

	/* get and decode payload from body */
	var payload GitlabPayload
	decoder := json.NewDecoder(reader.Body)
	err := decoder.Decode(&payload)
	if err != nil {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		Lg(1, "400: %s - %s (Failed to decode Payload)\n", reader.Method, reader.URL)
		return
	}

	message := queueMessageFromPayload(payload)

	/* publish message */
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

type GitlabPayload struct {
	ObjectKind   string `json:"object_kind"`
	Before       string `json:"before"`
	After        string `json:"after"`
	Ref          string `json:"ref"`
	CheckoutSha  string `json:"checkout_sha"`
	UserID       int    `json:"user_id"`
	UserName     string `json:"user_name"`
	UserUsername string `json:"user_username"`
	UserEmail    string `json:"user_email"`
	UserAvatar   string `json:"user_avatar"`
	ProjectID    int    `json:"project_id"`
	Project      struct {
		Name              string      `json:"name"`
		Description       string      `json:"description"`
		WebURL            string      `json:"web_url"`
		AvatarURL         interface{} `json:"avatar_url"`
		GitSSHURL         string      `json:"git_ssh_url"`
		GitHTTPURL        string      `json:"git_http_url"`
		Namespace         string      `json:"namespace"`
		VisibilityLevel   int         `json:"visibility_level"`
		PathWithNamespace string      `json:"path_with_namespace"`
		DefaultBranch     string      `json:"default_branch"`
		Homepage          string      `json:"homepage"`
		URL               string      `json:"url"`
		SSHURL            string      `json:"ssh_url"`
		HTTPURL           string      `json:"http_url"`
	} `json:"project"`
	Repository struct {
		Name            string `json:"name"`
		URL             string `json:"url"`
		Description     string `json:"description"`
		Homepage        string `json:"homepage"`
		GitHTTPURL      string `json:"git_http_url"`
		GitSSHURL       string `json:"git_ssh_url"`
		VisibilityLevel int    `json:"visibility_level"`
	} `json:"repository"`
	Commits []struct {
		ID        string    `json:"id"`
		Message   string    `json:"message"`
		Timestamp time.Time `json:"timestamp"`
		URL       string    `json:"url"`
		Author    struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"author"`
		Added    []string      `json:"added"`
		Modified []string      `json:"modified"`
		Removed  []interface{} `json:"removed"`
	} `json:"commits"`
	TotalCommitsCount int `json:"total_commits_count"`
}
