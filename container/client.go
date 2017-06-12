package container

import (
	"context"
	"encoding/json"
	"io"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerClient interface
type DockerClient interface {
	BuildImage(ctx context.Context, cloneURL, ref, name, tag string, notify BuildNotify) error
	Info(ctx context.Context) (string, error)
}

type BuildNotify interface {
	SendBuildReport(ctx context.Context, readCloser io.ReadCloser, target BuildTarget)
}

type BuildTarget struct {
	Name       string
	Tag        string
	GitContext string
}

// NewClient returns a new Client instance which can be used to interact with
// the Docker API.
func NewClient() DockerClient {
	apiClient, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Error instantiating Docker engine-api: %s", err)
	}

	return dockerAPI{apiClient: apiClient}
}

type dockerAPI struct {
	apiClient *client.Client
}

func (api dockerAPI) BuildImage(ctx context.Context, cloneURL, ref, name, tag string, notify BuildNotify) error {
	// set build options
	var options types.ImageBuildOptions
	options.RemoteContext = cloneURL + "#" + ref
	options.ForceRemove = true
	options.Tags = []string{name + ":" + tag}
	log.Debugf("Building Docker image with options: %+v", options)
	// execute build
	buildResponse, err := api.apiClient.ImageBuild(ctx, nil, options)
	// get build error
	if err != nil {
		log.Error(err)
		return err
	}
	// set build target
	var buildTarget BuildTarget
	buildTarget.GitContext = options.RemoteContext
	buildTarget.Name = name
	buildTarget.Tag = tag

	// send output and status
	notify.SendBuildReport(ctx, buildResponse.Body, buildTarget)
	return nil
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
