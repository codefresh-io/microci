package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	"gopkg.in/go-playground/webhooks.v3"
	"gopkg.in/go-playground/webhooks.v3/github"

	"github.com/codefresh-io/microci/container"
)

const (
	gitHubPath = "/github"
)

var (
	gClient         container.DockerClient
	gStopChan       chan bool
	gSlackToken     string
	gSlackChannel   string
	gCancelCommands []interface{}
	gNotify         container.BuildNotify
)

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
					Name:        "slack-token",
					Usage:       "Slack API token",
					Destination: &gSlackToken,
				},
				cli.StringFlag{
					Name:        "slack-channel",
					Usage:       "Slack channel to post build results",
					Destination: &gSlackChannel,
				},
			},
			Usage:       "start webhook server",
			ArgsUsage:   "configuration file",
			Description: "start webhook server to handle webhook events from GitHub",
			Action:      webhookServer,
			Before:      beforeCommand,
		},
		{
			Name:        "info",
			Usage:       "docker info",
			Description: "show docker info",
			Action:      dockerInfo,
			Before:      beforeCommand,
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
	// Set-up container client
	tls, err := tlsConfig(c)
	if err != nil {
		return err
	}

	// create new Docker client
	gClient = container.NewClient(c.GlobalString("host"), tls)

	// set global Notify object to Slack or STDOUT
	if gSlackToken != "" {
		gNotify = SlackNotify{}
	} else {
		gNotify = StdoutNotify{}
	}

	// handle stop signals
	handleSignals()
	return nil
}

func handleSignals() {
	// Graceful shut-down on SIGINT/SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// channel to notify long running commands to stop and cleanup
	// long running commands must listen to this channel and react
	gStopChan = make(chan bool, 1)

	go func() {
		sid := <-sigs
		log.Debugf("Received signal: %d", sid)
		gStopChan <- true
		for _, cancelFn := range gCancelCommands {
			log.Debug("Canceling running command")
			cancelFn.(context.CancelFunc)()
		}
		fmt.Println("Graceful exit :-)")
		os.Exit(0)
	}()
}

// tlsConfig translates the command-line options into a tls.Config struct
func tlsConfig(c *cli.Context) (*tls.Config, error) {
	var tlsConfig *tls.Config
	var err error
	caCertFlag := c.GlobalString("tlscacert")
	certFlag := c.GlobalString("tlscert")
	keyFlag := c.GlobalString("tlskey")

	if c.GlobalBool("tls") || c.GlobalBool("tlsverify") {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: !c.GlobalBool("tlsverify"),
		}

		// Load CA cert
		if caCertFlag != "" {
			var caCert []byte
			if strings.HasPrefix(caCertFlag, "/") {
				caCert, err = ioutil.ReadFile(caCertFlag)
				if err != nil {
					return nil, err
				}
			} else {
				caCert = []byte(caCertFlag)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}

		// Load client certificate
		if certFlag != "" && keyFlag != "" {
			var cert tls.Certificate
			if strings.HasPrefix(certFlag, "/") && strings.HasPrefix(keyFlag, "/") {
				cert, err = tls.LoadX509KeyPair(certFlag, keyFlag)
				if err != nil {
					return nil, err
				}
			} else {
				cert, err = tls.X509KeyPair([]byte(certFlag), []byte(keyFlag))
				if err != nil {
					return nil, err
				}
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}
	return tlsConfig, nil
}

// beforeCommand run before each command
func beforeCommand(c *cli.Context) error {
	// get recurrent time interval
	return nil
}

// Serve webhooks
func webhookServer(c *cli.Context) {
	// get GitHub secret
	secret := c.String("secret")
	// get port
	port := c.Int("port")
	// print status
	fmt.Printf("Listening for GitHub hooks on port: %d ...\n", port)
	// create new webhook
	hook := github.New(&github.Config{Secret: secret})
	// register push event handler
	hook.RegisterEvents(handlePushEvent, github.PushEvent)
	// register create event handler
	hook.RegisterEvents(handleCreateEvent, github.CreateEvent)
	// start webhook server
	err := webhooks.Run(hook, ":"+strconv.Itoa(port), gitHubPath)
	if err != nil {
		log.Error(err)
	}
}

func dockerInfo(c *cli.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	gCancelCommands = append(gCancelCommands, cancel)
	info, err := gClient.Info(ctx)
	if err != nil {
		log.Error(err)
	}
	fmt.Printf(info)
}
