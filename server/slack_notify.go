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

	// format build report message
	var output string
	output = fmt.Sprintf("*Build:* `%s:%s`\n", target.Name, target.Tag)
	output += fmt.Sprintf("*git context:* %s\n", target.GitContext)
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
	output += fmt.Sprintf("*build duration:* %s\n", time.Since(start))
	log.Debugf("Build %s:%s completed", target.Name, target.Tag)

	// prepare Slack message
	api := slack.New(gSlackToken)
	params := slack.PostMessageParameters{}
	params.IconEmoji = ":whale:"
	params.Markdown = true
	params.Username = "microci"

	// post Slack message
	channelID, timestamp, err := api.PostMessage(gSlackChannel, output, params)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	log.Debugf("Message successfully sent to channel %s at %s", channelID, timestamp)
}
