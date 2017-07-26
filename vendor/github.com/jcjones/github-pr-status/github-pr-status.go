// Copyright 2015 ISRG.  All rights reserved
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/codegangsta/cli"
	"github.com/jcjones/github-pr-status/api"
)

var authenticationConfig api.AuthConfig

func showError(c *cli.Context, err error) error {
	fmt.Println("ERROR: ")
	fmt.Println(err.Error())
	fmt.Println("")
	cli.ShowCommandHelp(c, c.Command.Name)
	return err
}

func actionSetStatus(c *cli.Context) {
	var err error

	if !c.IsSet("state") {
		err = fmt.Errorf("A state must be specified")
		showError(c, err)
		return
	}
	state := c.String("state")

	switch state {
	case "pending":
		fallthrough
	case "success":
		fallthrough
	case "error":
		fallthrough
	case "failure":
		// OK
		break
	default:
		showError(c, fmt.Errorf("Not a valid state: %s", c.String("state")))
		return
	}

	if !c.IsSet("context") {
		err = fmt.Errorf("A context must be specified")
		showError(c, err)
		return
	}
	context := c.String("context")

	if !c.IsSet("sha") {
		err = fmt.Errorf("A commit ID must be specified")
		showError(c, err)
		return
	}
	sha := c.String("sha")

	// Optionals
	url := c.String("url")
	description := c.String("description")

	// Accept description from stdin
	if description == "-" {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			showError(c, err)
			return
		}

		description = string(data)
	}

	// Checked at the global scope
	owner := c.GlobalString("owner")
	repo := c.GlobalString("repo")
	debug := c.GlobalBool("debug")

	err = api.StatusSet(authenticationConfig, owner, repo, sha, state, context, url, description, debug)
	if err != nil {
		fmt.Printf("Could not post: %s\n", err.Error())
	}
}

func actionAddComment(c *cli.Context) {
	var err error

	if !c.IsSet("body") {
		err = fmt.Errorf("A comment body must be specified")
		showError(c, err)
		return
	}
	body := c.String("body")

	// Accept body from stdin
	if body == "-" {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			showError(c, err)
			return
		}

		body = string(data)
	}

	if !c.IsSet("pr") {
		err = fmt.Errorf("A PR must be specified")
		showError(c, err)
		return
	}

	pr := c.Int("pr")
	if pr < 1 {
		err = fmt.Errorf("PRs must be >0")
		showError(c, err)
		return
	}

	fmt.Printf("Comment PR=#%d\n", pr)

	// Checked at the global scope
	owner := c.GlobalString("owner")
	repo := c.GlobalString("repo")
	debug := c.GlobalBool("debug")

	err = api.IssueAddComment(authenticationConfig, owner, repo, body, pr, debug)
	if err != nil {
		fmt.Printf("Could not post: %s\n", err.Error())
	}

}

func loadAuthDetailsAndConfigureLogging(c *cli.Context) (err error) {
	if !c.GlobalIsSet("authfile") {
		err = fmt.Errorf("You must specifiy a path to the authfile.")
		return showError(c, err)
	}

	path := c.GlobalString("authfile")
	configJSON, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Could not read authentication details. Path=%s Err=%s", path, err)
		return showError(c, err)
	}

	err = json.Unmarshal(configJSON, &authenticationConfig)
	if err != nil {
		err = fmt.Errorf("Could not parse authentication details. Err=%s", err)
		return showError(c, err)
	}

	if authenticationConfig.Type == nil || *authenticationConfig.Type != "basic" {
		err = fmt.Errorf("Only authentication type supported is 'basic'.")
		return showError(c, err)
	}

	if authenticationConfig.Username == nil {
		err = fmt.Errorf("You must provide a username in the authfile.")
		return showError(c, err)
	}

	if authenticationConfig.Password == nil {
		err = fmt.Errorf("You must provide a password or token in the authfile.")
		return showError(c, err)
	}

	if !c.GlobalIsSet("owner") {
		err = fmt.Errorf("You must provide a repo owner.")
		return showError(c, err)
	}

	if !c.GlobalIsSet("repo") {
		err = fmt.Errorf("You must provide a repo to operate on.")
		return showError(c, err)
	}

	return
}

func main() {
	app := cli.NewApp()
	app.Usage = "GitHub CI PR Tools"
	app.Version = "0.4.0"
	app.Copyright = "Copyright 2015 ISRG.  All rights reserved."
	app.Authors = []cli.Author{
		cli.Author{Name: "James 'J.C.' Jones", Email: "jcjones@letsencrypt.org"},
	}
	app.Before = loadAuthDetailsAndConfigureLogging

	app.Flags = []cli.Flag{
		cli.BoolFlag{Name: "debug, D", Usage: "Enable debugging output"},
		cli.StringFlag{Name: "owner, o", Usage: "Owner segment of GitHub Path"},
		cli.StringFlag{Name: "repo, r", Usage: "Repo segment of GitHub Path"},
		cli.StringFlag{Name: "authfile, f", Usage: "Authentication details"},
	}

	statusCommand := cli.Command{
		Name:   "status",
		Usage:  "Set or update a status on a commit",
		Action: actionSetStatus,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "sha, i", Usage: "Commit ID"},
			cli.StringFlag{Name: "state, s", Usage: "One of {pending, success, error, failure}"},
			cli.StringFlag{Name: "url, l", Usage: "URL to link from the status"},
			cli.StringFlag{Name: "context, c", Usage: "Status context, such as 'integration/gofmt'"},
			cli.StringFlag{Name: "description, d", Usage: "Extra descriptive text, or '-' to read from stdin"},
		},
	}

	commentCommand := cli.Command{
		Name:   "comment",
		Usage:  "Add a comment to a PR",
		Action: actionAddComment,
		Flags: []cli.Flag{
			cli.IntFlag{Name: "pr, n", Usage: "Pull Request #"},
			cli.StringFlag{Name: "body, b", Usage: "Body of the comment, or '-' to read from stdin"},
		},
	}

	app.Commands = []cli.Command{
		statusCommand, commentCommand,
	}

	app.Run(os.Args)
}
