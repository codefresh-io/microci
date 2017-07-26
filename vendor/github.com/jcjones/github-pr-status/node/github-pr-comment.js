// Copyright 2015 ISRG.  All rights reserved
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Posts a comment to Github on a PR.
//
// Example usage:
// echo "hi" | node github-pr-comment.js -f ./github-secret.json -o jcjones \
//   -r github-pr-status -n 1 -D
//
"use strict";

var Client = require("github");
var stdio = require('stdio');

var ops = stdio.getopt({
    'owner': {key: 'o', args: 1, description: 'Owner segment of Github repo name', mandatory: true},
    'repo': {key: 'r', args: 1, description: 'Github repo name', mandatory: true},
    'pr': {key: 'n', args: 1, description: 'PR ID'},
    'authfile': {key: 'f', args: 1, description: 'authentication file', mandatory: true},
    'debug': {key: 'D', description: 'enable debugging output'},
});

var github = new Client({
    debug: ops.debug,
    version: "3.0.0"
});

var authfile = require(ops.authfile);
github.authenticate(authfile);

function clone(a) {
   return JSON.parse(JSON.stringify(a));
}

function showResult(err, res){
  if (err != null) {
    console.log("Error: " + err);
    return;
  }

  if (ops.debug) {
    console.log("Result: " + res)
  }
}


function showDetails(err, res) {
  if (err != null) {
    console.log("Error: " + err);
    return;
  }

  console.log(res)

  if (ops.debug) {
    res.forEach(function (data){
      console.log("DATA: " + data);
    });
  }
}

if (ops.debug) {
  console.log("Configuration: " + JSON.stringify(ops));
}

stdio.read(function(text){
  var data = {};
  data["number"] = ops.pr;
  data["user"] = ops.owner;
  data["repo"] = ops.repo;
  data["body"] = text

  // github.issues.getComments(data, showDetails);

  github.issues.createComment(data, showResult);
});




