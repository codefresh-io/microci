package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// StatusUnknown unknown build status
	StatusUnknown = iota
	// StatusPassed passed build status
	StatusPassed
	// StatusRunning running build status
	StatusRunning
	// StatusFailed failed build status
	StatusFailed
	// StatusError error build status
	StatusError
)

var statusText = []string{"Unknown", "Passed", "Running", "Failed", "Error"}

// BuildReport build details
type BuildReport struct {
	RepoName     string
	Owner        string
	Tag          string
	ImageName    string
	BuildContext string
	Duration     time.Duration
	Start        time.Time
	// status should be set through function
	status int
	// status chang notification
	StatusNotify GitStatusNotify
}

// SetStatus set build report status and update Git repository
func (r *BuildReport) SetStatus(status int) {
	r.status = status % 5
	if r.StatusNotify != nil {
		go r.StatusNotify.UpdateStatus(r.Owner, r.RepoName, r.Tag, r.status)
	}
}

// GetStatus return status text
func (r *BuildReport) GetStatus() string {
	return statusText[r.status]
}

// BuildStats all build reports for current MicroCI instance
type BuildStats struct {
	reports []BuildReport
}

// SendReport send new build report to gStats channel
func (s *BuildStats) SendReport(report BuildReport) {
	s.reports = append(s.reports, report)
}

// GetStatsReport string with all builds fro current MicroCI instance
func (s *BuildStats) GetStatsReport() string {
	var report string
	for _, r := range s.reports {
		repJSON, err := json.Marshal(r)
		if err != nil {
			break
		}
		report += fmt.Sprintln(string(repJSON))
	}
	return report
}

// ReportHandler HTTP handler helper function
func (s *BuildStats) ReportHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "MicroCI Build Reports")
	fmt.Fprintln(w, "=====================")
	fmt.Fprintln(w, s.GetStatsReport())
}
