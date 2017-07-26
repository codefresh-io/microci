// Copyright 2015 ISRG.  All rights reserved
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const apiHost = "https://api.github.com"

// AuthConfig defines a user and password for the service
type AuthConfig struct {
	Type     *string
	Username *string
	Password *string
}

func post(authenticationConfig AuthConfig, debug bool, url string, postBody interface{}) error {
	client := &http.Client{}

	jsonStr, err := json.Marshal(postBody)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", jsonStr)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(*authenticationConfig.Username, *authenticationConfig.Password)

	if debug {
		fmt.Println("request Header:", req.Header)
		fmt.Println("request Body:", string(jsonStr))
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	contents, _ := ioutil.ReadAll(resp.Body)

	if debug {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		fmt.Println("response Body:", string(contents))
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("%s %s", resp.Status, string(contents))
	}

	return nil
}

// StatusSet sets the state of a given commit ID
func StatusSet(authenticationConfig AuthConfig, owner, repo, commitID, state, context, url, description string, debug bool) error {
	postBody := struct {
		State       string `json:"state"`
		Context     string `json:"context"`
		URL         string `json:"target_url,omitempty"`
		Description string `json:"description,omitempty"`
	}{
		State:       state,
		Context:     context,
		URL:         url,
		Description: description,
	}

	apiURL := fmt.Sprintf("%s/repos/%s/%s/statuses/%s", apiHost, owner, repo, commitID)

	return post(authenticationConfig, debug, apiURL, postBody)
}

// IssueAddComment adds a comment to the provided issue #
func IssueAddComment(authenticationConfig AuthConfig, owner, repo, body string, issueID int, debug bool) error {
	postBody := struct {
		Body string `json:"body"`
	}{
		Body: body,
	}

	apiURL := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", apiHost, owner, repo, issueID)

	return post(authenticationConfig, debug, apiURL, postBody)
}
