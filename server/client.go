package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// DockerClient interface
type DockerClient interface {
	RegistryLogin(ctx context.Context, user, password, registry string) error
	BuildPushImage(ctx context.Context, cloneURL, ref, name, fullname, tag string, notify BuildNotify) error
	Info(ctx context.Context) (string, error)
}

// BuildNotify interface
type BuildNotify interface {
	SendBuildReport(ctx context.Context, readCloser io.ReadCloser, target BuildTarget)
	SendPushReport(ctx context.Context, readCloser io.ReadCloser, image string)
}

// BuildTarget build target details
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

	return &dockerAPI{apiClient: apiClient}
}

type dockerAPI struct {
	apiClient  *client.Client
	authBase64 string
}

// Login to DockerRegistry
func (api *dockerAPI) RegistryLogin(ctx context.Context, user, password, registry string) error {
	auth := types.AuthConfig{
		Username: user,
		Password: password,
	}
	if registry != "" {
		auth.ServerAddress = registry
	}
	authBytes, _ := json.Marshal(auth)
	api.authBase64 = base64.URLEncoding.EncodeToString(authBytes)
	_, err := api.apiClient.RegistryLogin(ctx, auth)
	return err
}

func (api *dockerAPI) BuildPushImage(ctx context.Context, cloneURL, ref, name, fullname, tag string, notify BuildNotify) error {
	// set build options
	var options types.ImageBuildOptions
	options.RemoteContext = cloneURL + "#" + ref
	options.ForceRemove = true
	// create name for image to build
	var imageName string
	if gRegistry != "" {
		imageName += gRegistry + "/"
	}
	if gRepository != "" {
		imageName += gRepository + "/" + name
	} else {
		imageName += fullname
	}
	// get branch fro ref (if branch) or tag
	refs := strings.Split(ref, "/")
	tagText := refs[len(refs)-1]
	// prepare 2 tags
	// - one tag with commit
	// - one tag with branch
	options.Tags = []string{imageName + ":" + tagText, imageName + ":" + tag}
	// debug build options
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

	// send build output and status
	notify.SendBuildReport(ctx, buildResponse.Body, buildTarget)

	// push new image
	pushOptions := types.ImagePushOptions{}
	pushOptions.RegistryAuth = api.authBase64
	for _, image := range options.Tags {
		pushResponse, err := api.apiClient.ImagePush(ctx, image, pushOptions)
		// get push error
		if err != nil {
			log.Error(err)
			return err
		}
		// send output and status
		notify.SendPushReport(ctx, pushResponse, image)
	}
	return nil
}

// Info get Docker info
func (api *dockerAPI) Info(ctx context.Context) (string, error) {
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
