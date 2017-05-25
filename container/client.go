package container

import (
	"crypto/tls"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	engineapi "github.com/docker/docker/client"
)

const (
	// DefaultAPIVersion default Docker API version (Remote API v1.29 == Docker v17.05)
	DefaultAPIVersion = "v1.29"
)

// Client interface
type Client interface {
}

// NewClient returns a new Client instance which can be used to interact with
// the Docker API.
func NewClient(dockerHost string, tlsConfig *tls.Config) Client {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	verStr := DefaultAPIVersion
	if tmpStr := os.Getenv("DOCKER_API_VERSION"); tmpStr != "" {
		verStr = tmpStr
	}
	apiClient, err := engineapi.NewClient(dockerHost, verStr, client, nil)
	if err != nil {
		log.Fatalf("Error instantiating Docker engine-api: %s", err)
	}

	return dockerClient{apiClient: apiClient}
}

type dockerClient struct {
	apiClient engineapi.ContainerAPIClient
}
