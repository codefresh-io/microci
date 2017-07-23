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
	Notify         BuildNotify
	CancelCommands *[]interface{}
}

// New creates and returns a GitHubHook instance
func New(registry, repository string, commands *[]interface{}, notify BuildNotify, config *github.Config) *GitHubHook {
	return &GitHubHook{
		webhook:        github.New(config),
		CancelCommands: commands,
		Notify:         notify,
		Registry:       registry,
		Repository:     repository,
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
	*hook.CancelCommands = append(*hook.CancelCommands, cancel)
	go gClient.BuildPushImage(ctx, cloneURL, ref, pl.Repository.Name, pl.Repository.FullName, pl.HeadCommit.ID, hook.Registry, hook.Repository, hook.Notify)
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
		*hook.CancelCommands = append(*hook.CancelCommands, cancel)
		go gClient.BuildPushImage(ctx, cloneURL, ref, pl.Repository.Name, pl.Repository.FullName, ref, hook.Registry, hook.Repository, hook.Notify)
	}
}
