package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

// SlackNotify Slack notify interface
type SlackNotify struct {
}

// SendBuildReport send build output to slack channel
func (s SlackNotify) SendBuildReport(ctx context.Context, r io.ReadCloser, target BuildTarget) {
	defer r.Close()

	// create build report
	var buildReport BuildReport
	buildReport.BuildTarget = target
	// format build report message
	var output string
	output = fmt.Sprintf("*Docker Build:* `%s:%s`\n", target.Name, target.Tag)
	output += fmt.Sprintf("*git context:* %s\n", target.GitContext)
	buildReport.Start = time.Now()
	output += "```\n"
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		var line Line
		if err := json.Unmarshal([]byte(s), &line); err != nil {
			log.Error(err)
			return
		}
		output += fmt.Sprint(line.Stream)
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
	output += "```\n"
	buildReport.Duration = time.Since(buildReport.Start)
	// TODO: decide on build status
	buildReport.Status = "Completed"
	output += fmt.Sprintf("*build duration:* %s\n", buildReport.Duration)
	log.Debugf("Build %s:%s completed", target.Name, target.Tag)

	// send build report stats
	gStats.SendReport(buildReport)

	// prepare Slack message
	api := slack.New(gSlackToken)
	params := slack.PostMessageParameters{}
	params.IconEmoji = ":whale:"
	params.Markdown = true
	params.Username = "microci"

	// post Slack message
	channelID, timestamp, err := api.PostMessage(gSlackChannel, output, params)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Message successfully sent to channel %s at %s", channelID, timestamp)
}

// SendPushReport send push report
func (s SlackNotify) SendPushReport(ctx context.Context, r io.ReadCloser, image string) {
	defer r.Close()

	// format push report message
	output := fmt.Sprintf("*Docker Push:* `%s`\n", image)
	start := time.Now()
	output += "```\n"
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		var line Line
		if err := json.Unmarshal([]byte(s), &line); err != nil {
			log.Error(err)
			return
		}
		output += fmt.Sprint(line.Stream)
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
	output += "```\n"
	output += fmt.Sprintf("*push duration:* %s\n", time.Since(start))
	log.Debugf("Push %s completed", image)

	// prepare Slack message
	api := slack.New(gSlackToken)
	params := slack.PostMessageParameters{}
	params.IconEmoji = ":whale:"
	params.Markdown = true
	params.Username = "microci"

	// post Slack message
	channelID, timestamp, err := api.PostMessage(gSlackChannel, output, params)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
