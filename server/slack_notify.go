package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/nlopes/slack"
)

// SlackNotify Slack notify interface
type SlackNotify struct {
	token   string
	channel string
}

// SendBuildReport send build output to slack channel
func (s SlackNotify) SendBuildReport(ctx context.Context, r io.ReadCloser, target BuildTarget) {
	defer r.Close()

	// create build report
	var buildReport BuildReport
	buildReport.BuildTarget = target
	// format build report message
	var output []string
	buildReport.Start = time.Now()
	// regexp to decide on build status: FAILED by default
	re := regexp.MustCompile("Successfully built ([0-9a-f]{12})")
	buildReport.Status = "FAILED"
	// prepare output
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		var line Line
		if err := json.Unmarshal([]byte(s), &line); err != nil {
			log.Error(err)
			return
		}
		output = append(output, line.Stream)
		status := re.FindString(line.Stream)
		if status != "" {
			buildReport.Status = "PASSED"
		}
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
	buildReport.Duration = time.Since(buildReport.Start)

	// send build report stats
	gStats.SendReport(buildReport)

	// prepare Slack message
	api := slack.New(s.token)
	params := slack.PostMessageParameters{}
	params.IconEmoji = ":whale:"
	params.Markdown = true
	params.Username = "microci"

	// prepare attachment
	attachment := slack.Attachment{
		Pretext: "*New Docker build report from MicroCI*",
		Text:    strings.Join(output, "\n"),
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "Status",
				Value: buildReport.Status,
			},
			slack.AttachmentField{
				Title: "Duration",
				Value: buildReport.Duration.String(),
			},
			slack.AttachmentField{
				Title: "Git Context",
				Value: target.GitContext,
			},
		},
	}
	params.Attachments = []slack.Attachment{attachment}

	// post Slack message
	channelID, timestamp, err := api.PostMessage(s.channel, "", params)
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
	start := time.Now()
	var output string
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

	// prepare Slack message
	api := slack.New(s.token)
	params := slack.PostMessageParameters{}
	params.IconEmoji = ":whale:"
	params.Markdown = true
	params.Username = "microci"

	// post Slack message
	attachment := slack.Attachment{
		Pretext: "*New Docker push report from MicroCI*",
		Text:    output,
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "Duration",
				Value: time.Since(start).String(),
			},
		},
	}
	params.Attachments = []slack.Attachment{attachment}
	channelID, timestamp, err := api.PostMessage(s.channel, "", params)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
