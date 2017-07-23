package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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
