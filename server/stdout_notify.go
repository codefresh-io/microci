package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Line single line text
type Line struct {
	Stream string `json:"stream"`
}

// StdoutNotify notify to STDOUT
type StdoutNotify struct {
}

// SendBuildReport stream build output to STDOUT
func (out StdoutNotify) SendBuildReport(ctx context.Context, r io.ReadCloser, target BuildTarget) {
	defer r.Close()
	// print build status
	fmt.Println("===== New Build =====")
	fmt.Printf("Building %s:%s\n", target.Name, target.Tag)
	fmt.Printf("From git context: %s\n", target.GitContext)
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
	fmt.Printf("Build duration: %s\n", time.Since(start))
	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
}
