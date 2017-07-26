package main

import (
	github "github.com/jcjones/github-pr-status/api"
)

// GitStatusNotify - report build status update per commit
type GitStatusNotify interface {
	UpdateStatus(owner, repo, sha string, status int) error
}

// GitHubStatusNotify report - report build status per GitHub commit/PR
type GitHubStatusNotify struct {
	authConfig     github.AuthConfig
	githubStatuses []string
}

// NewGitHubStatusNotify create new GitHubStatusReport
func NewGitHubStatusNotify(user, token string) *GitHubStatusNotify {
	var basic = "basic"
	authConfig := github.AuthConfig{Type: &basic, Username: &user, Password: &token}
	// Valid GitHub statuses: pending, success, error, failure
	statuses := []string{"pending", "success", "pending", "failure", "error"}
	return &GitHubStatusNotify{authConfig, statuses}
}

// ReportStatus report build status to GitHub
func (gh GitHubStatusNotify) UpdateStatus(owner, repo, sha string, status int) error {
	return github.StatusSet(gh.authConfig, owner, repo, sha, gh.githubStatuses[status%5], "continuous-integration/microci", "", "", false)
}
