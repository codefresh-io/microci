package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
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
	gClient   container.Client
	gStopChan chan bool
	gWG       sync.WaitGroup
	gWorkdir  string
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
	rootCertPath := "/etc/ssl/docker"

	if os.Getenv("DOCKER_CERT_PATH") != "" {
		rootCertPath = os.Getenv("DOCKER_CERT_PATH")
	}

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
					Name:        "workdir",
					Usage:       "working directory for git clone/pull operations",
					Value:       ".",
					Destination: &gWorkdir,
				},
			},
			Usage:       "start webhook server",
			ArgsUsage:   "configuration file",
			Description: "start webhook server to handle Push events coming from GitHub",
			Action:      webhookServer,
			Before:      beforeCommand,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host, H",
			Usage:  "daemon socket to connect to",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
		},
		cli.BoolFlag{
			Name:  "tls",
			Usage: "use TLS; implied by --tlsverify",
		},
		cli.BoolFlag{
			Name:   "tlsverify",
			Usage:  "use TLS and verify the remote",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
		cli.StringFlag{
			Name:  "tlscacert",
			Usage: "trust certs signed only by this CA",
			Value: fmt.Sprintf("%s/ca.pem", rootCertPath),
		},
		cli.StringFlag{
			Name:  "tlscert",
			Usage: "client certificate for TLS authentication",
			Value: fmt.Sprintf("%s/cert.pem", rootCertPath),
		},
		cli.StringFlag{
			Name:  "tlskey",
			Usage: "client key for TLS authentication",
			Value: fmt.Sprintf("%s/key.pem", rootCertPath),
		},
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
	// create new Chaos instance
	//chaos = action.NewChaos()
	// habdle termination signal

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
		log.Debug("Sending stop signal to running steps ...")
		gWG.Wait()
		log.Debug("Graceful exit :-)")
		os.Exit(1)
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
	// create new webhook
	hook := github.New(&github.Config{Secret: secret})
	// register event handler
	hook.RegisterEvents(handlePushEvent, github.PushEvent)
	// start webhook server
	err := webhooks.Run(hook, ":"+strconv.Itoa(port), gitHubPath)
	if err != nil {
		log.Error(err)
	}
}
