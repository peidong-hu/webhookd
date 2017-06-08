package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func queueMessageFromGithub(p GithubPayload) (m MQMessage) {
	m.Version = "0.0"
	m.Repository = p.Repository.FullName
	m.Branch = p.Repository.DefaultBranch
	m.Commit = p.HeadCommit.TreeID
	m.Message = p.HeadCommit.Message
	m.Author = p.HeadCommit.Author.Username
	m.Trigger = "GitHub Push"
	return m
}


func githubHandler(writer http.ResponseWriter, reader *http.Request) {

	/* check request type */
	if reader.Method != "POST" {
		/* 405 Method Not Allowed */
		writer.Header().Set("Allow", "POST")
		http.Error(writer, http.StatusText(405), 405)
		lg(1, "405: %s - %s\n", reader.Method, reader.URL)
		return
	}

	/* check GitHub headers */
	event := reader.Header.Get("X-GitHub-Event")
	delivery := reader.Header.Get("X-GitHub-Delivery")
	if event == "" || delivery == "" {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		lg(1, "400: %s - %s (Missing GitHub Header)\n", reader.Method, reader.URL)
		return
	}

	/* only process push-events */
	if event != "push" {
		/* thanks and goodbye */
		writer.WriteHeader(200)
		writer.Write([]byte("OK\n"))
		lg(1, "Ignoring Event %s for %s", event, reader.URL)
		return
	}

	/* check Content-Type header */
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

	/* verify signature */
	signature := reader.Header.Get("X-Hub-Signature")
	err = checkGithubSignature(rawPayload, signature, CONFIG.GithubSecret)
	if err != nil {
		/* 400 Bad Request */
		http.Error(writer, http.StatusText(400), 400)
		lg(1, "400: %s - %s (Invalid signature)\n", reader.Method, reader.URL)
		return
	}

	/* decode payload */
	payload := GithubPayload
	err = json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		http.Error(writer, http.StatusText(400), 400)
		lg(0, "400: %s - %s (Error decoding JSON: %s)\n", reader.Method, reader.URL, err)
		return
	}

	/* generate queue message */
	rawMessage := queueMessageFromGithub(payload)

	message, _ := json.Marshal(rawMessage)

	/* publish message */
	err = publishMessage(MQCHANNEL, string(message))
	if err != nil {
		http.Error(writer, http.StatusText(500), 500)
		lg(0, "500: %s - %s (Failed to publish message: %s)\n", reader.Method, reader.URL, err.Error)
		return
	}

	/* close HTTP stream */
	writer.WriteHeader(200)
	writer.Write([]byte("OK\n"))

	return
}

func checkGithubSignature(rawPayload string, signature string, secret string) (err error) {
	if secret == "" {
		return nil
	}

	if ! strings.HasPrefix(signature, "sha1=") {
		return fmt.Errorf("format")
	}

	signature = strings.TrimPrefix(signature, "sha1=")
	requestMAC, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}

	hash := hmac.New(sha1.New, []byte(secret))
		_, _ = hash.Write(rawPayload)
	expectedMAC := hash.Sum(nil)
	if ! hmac.Equal(requestMAC, expectedMAC) {
		return fmt.Errorf("invalid secret")
	}

	return nil
}


/* https://developer.github.com/v3/activity/events/types/#pushevent */
type GithubPayload struct {
	Ref string `json:"ref"`
	Before string `json:"before"`
	After string `json:"after"`
	Created bool `json:"created"`
	Deleted bool `json:"deleted"`
	Forced bool `json:"forced"`
	BaseRef interface{} `json:"base_ref"`
	Compare string `json:"compare"`
	Commits []struct {
		ID string `json:"id"`
		TreeID string `json:"tree_id"`
		Distinct bool `json:"distinct"`
		Message string `json:"message"`
		Timestamp string `json:"timestamp"`
		URL string `json:"url"`
		Author struct {
			Name string `json:"name"`
			Email string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Name string `json:"name"`
			Email string `json:"email"`
			Username string `json:"username"`
		} `json:"committer"`
		Added []interface{} `json:"added"`
		Removed []interface{} `json:"removed"`
		Modified []string `json:"modified"`
	} `json:"commits"`
	HeadCommit struct {
		ID string `json:"id"`
		TreeID string `json:"tree_id"`
		Distinct bool `json:"distinct"`
		Message string `json:"message"`
		Timestamp string `json:"timestamp"`
		URL string `json:"url"`
		Author struct {
			Name string `json:"name"`
			Email string `json:"email"`
			Username string `json:"username"`
		} `json:"author"`
		Committer struct {
			Name string `json:"name"`
			Email string `json:"email"`
			Username string `json:"username"`
		} `json:"committer"`
		Added []interface{} `json:"added"`
		Removed []interface{} `json:"removed"`
		Modified []string `json:"modified"`
	} `json:"head_commit"`
	Repository struct {
		ID int `json:"id"`
		Name string `json:"name"`
		FullName string `json:"full_name"`
		Owner struct {
			Name string `json:"name"`
			Email string `json:"email"`
		} `json:"owner"`
		Private bool `json:"private"`
		HTMLURL string `json:"html_url"`
		Description string `json:"description"`
		Fork bool `json:"fork"`
		URL string `json:"url"`
		ForksURL string `json:"forks_url"`
		KeysURL string `json:"keys_url"`
		CollaboratorsURL string `json:"collaborators_url"`
		TeamsURL string `json:"teams_url"`
		HooksURL string `json:"hooks_url"`
		IssueEventsURL string `json:"issue_events_url"`
		EventsURL string `json:"events_url"`
		AssigneesURL string `json:"assignees_url"`
		BranchesURL string `json:"branches_url"`
		TagsURL string `json:"tags_url"`
		BlobsURL string `json:"blobs_url"`
		GitTagsURL string `json:"git_tags_url"`
		GitRefsURL string `json:"git_refs_url"`
		TreesURL string `json:"trees_url"`
		StatusesURL string `json:"statuses_url"`
		LanguagesURL string `json:"languages_url"`
		StargazersURL string `json:"stargazers_url"`
		ContributorsURL string `json:"contributors_url"`
		SubscribersURL string `json:"subscribers_url"`
		SubscriptionURL string `json:"subscription_url"`
		CommitsURL string `json:"commits_url"`
		GitCommitsURL string `json:"git_commits_url"`
		CommentsURL string `json:"comments_url"`
		IssueCommentURL string `json:"issue_comment_url"`
		ContentsURL string `json:"contents_url"`
		CompareURL string `json:"compare_url"`
		MergesURL string `json:"merges_url"`
		ArchiveURL string `json:"archive_url"`
		DownloadsURL string `json:"downloads_url"`
		IssuesURL string `json:"issues_url"`
		PullsURL string `json:"pulls_url"`
		MilestonesURL string `json:"milestones_url"`
		NotificationsURL string `json:"notifications_url"`
		LabelsURL string `json:"labels_url"`
		ReleasesURL string `json:"releases_url"`
		CreatedAt int `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		PushedAt int `json:"pushed_at"`
		GitURL string `json:"git_url"`
		SSHURL string `json:"ssh_url"`
		CloneURL string `json:"clone_url"`
		SvnURL string `json:"svn_url"`
		Homepage interface{} `json:"homepage"`
		Size int `json:"size"`
		StargazersCount int `json:"stargazers_count"`
		WatchersCount int `json:"watchers_count"`
		Language interface{} `json:"language"`
		HasIssues bool `json:"has_issues"`
		HasDownloads bool `json:"has_downloads"`
		HasWiki bool `json:"has_wiki"`
		HasPages bool `json:"has_pages"`
		ForksCount int `json:"forks_count"`
		MirrorURL interface{} `json:"mirror_url"`
		OpenIssuesCount int `json:"open_issues_count"`
		Forks int `json:"forks"`
		OpenIssues int `json:"open_issues"`
		Watchers int `json:"watchers"`
		DefaultBranch string `json:"default_branch"`
		Stargazers int `json:"stargazers"`
		MasterBranch string `json:"master_branch"`
	} `json:"repository"`
	Pusher struct {
		Name string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Sender struct {
		Login string `json:"login"`
		ID int `json:"id"`
		AvatarURL string `json:"avatar_url"`
		GravatarID string `json:"gravatar_id"`
		URL string `json:"url"`
		HTMLURL string `json:"html_url"`
		FollowersURL string `json:"followers_url"`
		FollowingURL string `json:"following_url"`
		GistsURL string `json:"gists_url"`
		StarredURL string `json:"starred_url"`
		SubscriptionsURL string `json:"subscriptions_url"`
		OrganizationsURL string `json:"organizations_url"`
		ReposURL string `json:"repos_url"`
		EventsURL string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type string `json:"type"`
		SiteAdmin bool `json:"site_admin"`
	} `json:"sender"`
}
