# GitHub PR Status tools (Go & Node.js)

These tools programmatically make updates to Pull Requests at GitHub. They are
intended to be used with build scripts, and to authenticate using [personal access tokens](https://github.com/settings/tokens).

## Authentication Notes

The only permissions your [personal access token](https://github.com/settings/tokens)
needs is `public_repo` or `repo` (if not public).

When you get your token, construct a file like [examples/example-secret.json](https://github.com/jcjones/github-pr-status/blob/master/examples/example-secret.json),
where `username` is your account and `password` is the secret access token you
generated.

```json
{
    "type": "basic",
    "username": "yourGitHubUsername",
    "password": "longUnguessableSecretToken"
}
```

The only `type` supported is `basic`, and you need to include it.

# Installation
Installation is different whether you want the Node.js packages, or use Go.

## Node.js
`npm install github-pr-status`

## Go
`go get github.com/jcjones/github-pr-status`

# Usage

The Go binary can provide useful help if you run it with `--help`, such as:
```
$ github-pr-status --help

NAME:
   github-pr-status - GitHub CI PR Tools

USAGE:
   github-pr-status [global options] command [command options] [arguments...]

VERSION:
   0.4.0

AUTHOR(S):
   James 'J.C.' Jones <jcjones@letsencrypt.org>

COMMANDS:
   status   Set or update a status on a commit
   comment  Add a comment to a PR
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -D      Enable debugging output
   --owner, -o      Owner segment of GitHub Path
   --repo, -r       Repo segment of GitHub Path
   --authfile, -f   Authentication details
   --help, -h       show help
   --version, -v    print the version

COPYRIGHT:
   Copyright 2015 ISRG.  All rights reserved.
```

## Setting status

These are flags to give checkmarks / x's on a PR. They actually are logged
against a particular Commit ID.

For help about the particular status types, check out [Github v3 Statuses API](https://developer.github.com/v3/repos/statuses/#create-a-status)

Node.js:
```bash
node node_modules/github-pr-status/github-pr-status.js -f ./github-secret.json -o jcjones -r \
  github-pr-status -s failure -i "0e296b7443b91d125f5b51e2d81663bcae667864" \
  -c "node/happy"
```

Go:
```bash
github-pr-status -D -f ./github-secret.json -o jcjones -r github-pr-status status  -i f2a751961d857a3ab00e693cacadfb4393c329ee -s failure --context "integration/ok"
```

## Appending PR comments

`github-pr-comment.js` posts the text from `stdin` to a comment on the provided
PR.

Node.js:
```bash
echo "hi" | node node_modules/github-pr-status/github-pr-comment.js -f ./github-secret.json -o jcjones \
  -r github-pr-status -n 1 -D
```

Go:
```bash
echo "OK" | github-pr-status -D -f ./github-secret.json -o jcjones -r github-pr-status comment -n 1 -b -
```