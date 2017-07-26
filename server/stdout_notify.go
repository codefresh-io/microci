package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

// Line single line text
type Line struct {
	Stream string `json:"stream"`
}

// StdoutNotify notify to STDOUT
type StdoutNotify struct {
	stats BuildStats
}

// SendBuildReport stream build output to STDOUT
func (out StdoutNotify) SendBuildReport(ctx context.Context, r io.ReadCloser, buildReport BuildReport) {
	defer r.Close()
	// print build status
	fmt.Println("===== Docker Build =====")
	fmt.Printf("Building %s:%s\n", buildReport.ImageName, buildReport.Tag)
	fmt.Printf("From git context: %s\n", buildReport.BuildContext)
	buildReport.Start = time.Now()
	// regexp to decide on build status: FAILED by default
	re := regexp.MustCompile("Successfully built ([0-9a-f]{12})")
	buildReport.SetStatus(StatusRunning)
	// stream build output
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		var line Line
		if err := json.Unmarshal([]byte(s), &line); err != nil {
			log.Error(err)
			break
		}
		fmt.Print(line.Stream)
		status := re.FindString(line.Stream)
		if status != "" {
			buildReport.SetStatus(StatusPassed)
		}
	}
	if buildReport.status != StatusPassed {
		buildReport.SetStatus(StatusFailed)
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
		buildReport.SetStatus(StatusError)
	}
	// output build status
	fmt.Printf("Build status: %s\n", buildReport.GetStatus())
	// calculate duration
	buildReport.Duration = time.Since(buildReport.Start)
	// print duration
	fmt.Printf("Build duration: %s\n", buildReport.Duration)
	// send build report stats
	out.stats.SendReport(buildReport)
}

// SendPushReport print push details
func (out StdoutNotify) SendPushReport(ctx context.Context, r io.ReadCloser, image string) {
	defer r.Close()

	// print push status
	fmt.Println("===== Docker Push =====")
	fmt.Printf("Pushing %s ...\n", image)

	start := time.Now()
	// stream build output
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		s := scanner.Text()
		var line Line
		if err := json.Unmarshal([]byte(s), &line); err != nil {
			log.Error(err)
			break
		}
		fmt.Print(line.Stream)
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
	}

	// print duration
	fmt.Printf("Push duration: %s\n", time.Since(start))
}
