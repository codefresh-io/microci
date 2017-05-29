package main

import (
	log "github.com/Sirupsen/logrus"

	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"
)

// handlePushEvent handles GitHub push events
func handlePushEvent(payload interface{}, header webhooks.Header) {

	log.Debug("Handling Push Request")

	pl := payload.(github.PushPayload)

	// Do whatever you want from here...
	log.Debugf("%+v", pl)
}
