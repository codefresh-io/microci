package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"gopkg.in/go-playground/webhooks.v3/github"
)

const (
	gitHubPath = "/github"
)

// Concurrent slice with all cancel commands from contexes
var gCancelCommands = NewConcurrentSlice()

// Singleton Docker Client
var once sync.Once
var dockerClient DockerClient

var getDockerClient = func() DockerClient {
	once.Do(func() {
		dockerClient = NewClient()
	})
	return dockerClient
}

var (
	// Version that is passed on compile time through -ldflags
	Version = "built locally"

	// GitCommit that is passed on compile time through -ldflags
	GitCommit = "none"

	// GitBranch that is passed on compile time through -ldflags
	GitBranch = "none"

	// BuildTime that is passed on compile time through -ldflags
	BuildTime = "none"

	// HumanVersion is a human readable app version
	HumanVersion = fmt.Sprintf("%s - %.7s (%s) %s", Version, GitCommit, GitBranch, BuildTime)

	// ASCIILogo for MicroCI
	ASCIILogo = `
                   _ ___ 
  |\/| o  _ ._ _  /   |  
  |  | | (_ | (_) \_ _|_ 
                        
        .
       ":"
     ___:_____    |"\/"|
   ,'         \    \  /
   |  O        \___/  |
 ~^~^~^~^~^~^~^~^~^~^~^~^~
 `
)

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{})
}

func main() {
	app := cli.NewApp()
	app.Name = "MicroCI"
	app.Version = HumanVersion
	app.Usage = "Minimalistic CI tool for Docker"
	app.UsageText = fmt.Sprintf(`
MicroCI is a minimalistic Continuous Integration (CI) tool designed and tuned for Docker-based microservices.
%s
 MicroCI respects DOCKER environment variables:
   - DOCKER_HOST        - set the url to the docker serve (default unix:///var/run/docker.sock)
   - DOCKER_API_VERSION - set the version of the API to reach
   - DOCKER_CERT_PATH   - path to load the TLS certificates from
   - DOCKER_TLS_VERIFY  - enable or disable TLS verification, off by default
   
Copyright © Codefresh.io`, ASCIILogo)
	app.Before = before
	app.Commands = []cli.Command{
		{
			Name: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "secret",
					Usage: "GitHub webhook secret",
				},
				cli.StringFlag{
					Name:  "ip",
					Usage: "ip the webhook should serve hooks on",
					Value: "0.0.0.0",
				},
				cli.IntFlag{
					Name:  "port",
					Usage: "port the webhook should serve hooks on",
					Value: 9000,
				},
				cli.StringFlag{
					Name:  "user, u",
					Usage: "Docker Registry user",
				},
				cli.StringFlag{
					Name:  "password, p",
					Usage: "Docker Registry password",
				},
				cli.StringFlag{
					Name:  "registry",
					Usage: "Docker Registry (DockerHub by default)",
				},
				cli.StringFlag{
					Name:  "repository",
					Usage: "Docker Registry repository (default: VCS account name)",
				},
				cli.StringFlag{
					Name:  "slack-token, t",
					Usage: "Slack API token",
				},
				cli.StringFlag{
					Name:  "slack-channel, c",
					Usage: "Slack channel to post build results",
				},
			},
			Usage:       "start webhook server",
			ArgsUsage:   "configuration file",
			Description: "start webhook server to handle webhook events from GitHub",
			Action:      webhookServer,
		},
		{
			Name:        "info",
			Usage:       "docker info",
			Description: "show docker info",
			Action:      dockerInfo,
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug mode with verbose logging",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "do not execute commands, just log",
		},
		cli.BoolFlag{
			Name:  "json",
			Usage: "produce log in JSON format: Logstash and Splunk friendly",
		},
	}

	app.Run(os.Args)
}

func before(c *cli.Context) error {
	// set debug log level
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	// set log formatter to JSON
	if c.GlobalBool("json") {
		log.SetFormatter(&log.JSONFormatter{})
	}

	// handle stop signals
	handleSignals(make(chan os.Signal, 1), true)
	return nil
}

func handleSignals(sigs chan os.Signal, exitOnSignal bool) {
	// Graceful shut-down on SIGINT/SIGTERM
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sid := <-sigs
		log.Debugf("Received signal: %d", sid)
		if sid == syscall.SIGTERM || sid == syscall.SIGKILL {
			for item := range gCancelCommands.Iter() {
				log.Debug("Canceling running command")
				item.Value.(context.CancelFunc)()
			}
			fmt.Println("\nGracefully exiting :-)")
			if exitOnSignal {
				os.Exit(0)
			}
		}
	}()
}

// handle webhook - testable function
func handleWebhook(c *cli.Context) http.Handler {
	// get GitHub secret
	secret := c.String("secret")
	// get slack token and channel
	slackToken := c.String("slack-token")
	slackChannel := c.String("slack-channel")
	// build stats
	stats := BuildStats{}
	// set global Notify object to Slack or STDOUT
	var notify BuildNotify
	notify = StdoutNotify{stats}
	if slackToken != "" {
		notify = SlackNotify{slackToken, slackChannel, stats}
	}
	// login to registry if credentials are passed
	user := c.String("user")
	password := c.String("password")
	registry := c.String("registry")
	repository := c.String("repository")
	if user != "" && password != "" {
		ctx, cancel := context.WithCancel(context.Background())
		gCancelCommands.Append(cancel)
		getDockerClient().RegistryLogin(ctx, user, password, registry)
	}

	// create new webhook
	githubHook := New(registry, repository, gCancelCommands, notify, &github.Config{Secret: secret})
	// register push event handler
	githubHook.RegisterPushEvent()
	// register create event handler
	githubHook.RegisterCreateEvent()

	// create HTTP server
	srv := http.NewServeMux()

	// handle github webhooks
	srv.HandleFunc(gitHubPath, func(w http.ResponseWriter, r *http.Request) {
		githubHook.ParsePayload(w, r)
	})
	// handle github webhooks
	srv.HandleFunc("/microci"+gitHubPath, func(w http.ResponseWriter, r *http.Request) {
		githubHook.ParsePayload(w, r)
	})

	// handle stats
	srv.HandleFunc("/report", stats.ReportHandler)
	srv.HandleFunc("/microci/report", stats.ReportHandler)

	srv.HandleFunc("/", statusHandler)
	srv.HandleFunc("/microci/", statusHandler)

	return srv
}

// Serve webhooks
func webhookServer(c *cli.Context) {
	// get port
	port := c.Int("port")
	// run http server and handle webhook
	err := http.ListenAndServe(":"+strconv.Itoa(port), handleWebhook(c))
	if err != nil {
		log.Error(err)
	} else {
		// print status
		fmt.Printf("Listening for GitHub hooks on port: %d ...\n", port)
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "MicroCI version %s is up and running\n%s", HumanVersion, ASCIILogo)
	// debug request
	queryValues := r.URL.Query()
	if queryValues.Get("debug") == "true" {
		fmt.Fprint(w, "\n\n\n")
		fmt.Fprintln(w, "Debug request:")
		fmt.Fprintln(w, "==============")
		// Save a copy of this request for debugging.
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Fprintln(w, string(requestDump))
	}
}

func dockerInfo(c *cli.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	gCancelCommands.Append(cancel)
	info, err := getDockerClient().Info(ctx)
	if err != nil {
		log.Error(err)
	}
	fmt.Printf(info)
}
