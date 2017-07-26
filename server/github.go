package main

import (
	"context"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"

	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

// GitHubHook is a Webhook with additional properties
type GitHubHook struct {
	webhook        *github.Webhook
	Registry       string
	Repository     string
	Token          string
	User           string
	Notify         BuildNotify
	StatusNotify   GitStatusNotify
	CancelCommands *ConcurrentSlice
}

// NewGitHubHook creates and returns a GitHubHook instance
func NewGitHubHook(registry, repository string, commands *ConcurrentSlice, notify BuildNotify, statusNotify GitStatusNotify, user, token string, config *github.Config) *GitHubHook {
	return &GitHubHook{
		webhook:        github.New(config),
		CancelCommands: commands,
		Notify:         notify,
		StatusNotify:   statusNotify,
		Registry:       registry,
		Repository:     repository,
		User:           user,
		Token:          token,
	}
}

// RegisterPushEvent register push event
func (hook GitHubHook) RegisterPushEvent() {
	hook.webhook.RegisterEvents(hook.handlePushEvent, github.PushEvent)
}

// ParsePayload parse HTTP payload
func (hook GitHubHook) ParsePayload(w http.ResponseWriter, r *http.Request) {
	hook.webhook.ParsePayload(w, r)
}

// RegisterCreateEvent register push event
func (hook GitHubHook) RegisterCreateEvent() {
	hook.webhook.RegisterEvents(hook.handleCreateEvent, github.CreateEvent)
}

// handlePushEvent handles GitHub push events
func (hook GitHubHook) handlePushEvent(payload interface{}, header webhooks.Header) {

	log.Debug("Handling Push Request")

	// get playload for push event
	pl := payload.(github.PushPayload)
	// Do whatever you want from here...
	log.Debugf("%+v", pl)

	// get branch fro ref (if branch) or tag
	refs := strings.Split(pl.Ref, "/")
	ref := refs[len(refs)-1]
	// get clone URL
	cloneURL := pl.Repository.CloneURL

	// do build
	ctx, cancel := context.WithCancel(context.Background())
	hook.CancelCommands.Append(cancel)
	client := getDockerClient()
	go client.BuildPushImage(ctx, cloneURL, ref, pl.Repository.Name, pl.Repository.Owner.Name, pl.HeadCommit.ID, hook.Registry, hook.Repository, hook.Notify, hook.StatusNotify)
}

// handleCreateEvent handles GitHub create events
func (hook GitHubHook) handleCreateEvent(payload interface{}, header webhooks.Header) {

	log.Debug("Handling Create Request")

	// get playload for push event
	pl := payload.(github.CreatePayload)
	// Do whatever you want from here...
	log.Debugf("%+v", pl)

	// get ref type: branch or tag and build
	if pl.RefType == "branch" || pl.RefType == "tag" {
		ref := pl.Ref
		cloneURL := pl.Repository.CloneURL
		// build
		ctx, cancel := context.WithCancel(context.Background())
		hook.CancelCommands.Append(cancel)
		client := getDockerClient()
		go client.BuildPushImage(ctx, cloneURL, ref, pl.Repository.Name, pl.Repository.Owner.Login, ref, hook.Registry, hook.Repository, hook.Notify, hook.StatusNotify)
	}
}
