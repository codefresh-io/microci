package main

import (
	"encoding/json"
	"fmt"
	"time"
)

var gStats BuildStats

// BuildReport build details
type BuildReport struct {
	BuildTarget
	Duration time.Duration
	Status   string
	Start    time.Time
}

// BuildStats all build reports for current MicroCI instance
type BuildStats struct {
	reports []BuildReport
	channel chan BuildReport
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
