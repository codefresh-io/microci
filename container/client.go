package container

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	// DefaultAPIVersion default Docker API version (Remote API v1.29 == Docker v17.05)
	DefaultAPIVersion = "v1.29"
)

// DockerClient interface
type DockerClient interface {
	BuildImage(ctx context.Context, cloneURL, ref, name, tag string) (io.ReadCloser, error)
	Info(ctx context.Context) (string, error)
}

// NewClient returns a new Client instance which can be used to interact with
// the Docker API.
func NewClient(dockerHost string, tlsConfig *tls.Config) DockerClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	verStr := DefaultAPIVersion
	if tmpStr := os.Getenv("DOCKER_API_VERSION"); tmpStr != "" {
		verStr = tmpStr
	}
	defaultHeaders := map[string]string{"User-Agent": "microci"}
	httpClient = nil
	apiClient, err := client.NewClient(dockerHost, verStr, httpClient, defaultHeaders)
	if err != nil {
		log.Fatalf("Error instantiating Docker engine-api: %s", err)
	}

	return dockerAPI{apiClient: apiClient}
}

type dockerAPI struct {
	apiClient *client.Client
}

func (api dockerAPI) BuildImage(ctx context.Context, cloneURL, ref, name, tag string) (io.ReadCloser, error) {
	// set build options
	var options types.ImageBuildOptions
	options.RemoteContext = cloneURL + "#" + ref
	options.ForceRemove = true
	options.Tags = []string{name + ":" + tag}
	log.Debugf("Building Docker image with options: %+v", options)
	// execute build
	buildResponse, err := api.apiClient.ImageBuild(ctx, nil, options)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return nil, err
	}
	response, err := ioutil.ReadAll(buildResponse.Body)
	if err != nil {
		fmt.Printf("%s", err.Error())
	}
	fmt.Println(string(response))
	return buildResponse.Body, nil
}

func (api dockerAPI) Info(ctx context.Context) (string, error) {
	info, err := api.apiClient.Info(ctx)
	if err != nil {
		return "", err
	}
	jsonInfo, err := json.Marshal(&info)
	if err != nil {
		return "", err
	}
	return string(jsonInfo), nil
}
