package main

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

// DockerClient interface
type DockerClient interface {
	RegistryLogin(ctx context.Context, user, password, registry string) error
	BuildPushImage(ctx context.Context, cloneURL, ref, repoName, owner, tag, registry, repository string, notify BuildNotify, statusNotify GitStatusNotify) error
	Info(ctx context.Context) (string, error)
}

// BuildNotify interface
type BuildNotify interface {
	SendBuildReport(ctx context.Context, readCloser io.ReadCloser, report BuildReport)
	SendPushReport(ctx context.Context, readCloser io.ReadCloser, image string)
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

// DockerClientAPI - a subset of Docker API used by MicroCI; wrap it with interface for testing/mocking
type DockerClientAPI interface {
	Info(ctx context.Context) (types.Info, error)
	RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error)
	ImagePush(ctx context.Context, ref string, options types.ImagePushOptions) (io.ReadCloser, error)
	ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
}

type dockerAPI struct {
	apiClient  DockerClientAPI
	authBase64 string
}

// Login to DockerRegistry
func (api *dockerAPI) RegistryLogin(ctx context.Context, user, password, registry string) error {
	if user != "" && password != "" {
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
	return nil
}

func (api *dockerAPI) BuildPushImage(ctx context.Context, cloneURL, ref, repoName, owner, tag, registry, repository string, notify BuildNotify, statusNotify GitStatusNotify) error {
	// set build options
	var options types.ImageBuildOptions
	options.RemoteContext = cloneURL + "#" + ref
	options.ForceRemove = true
	// create name for image to build
	var imageName string
	if registry != "" {
		imageName += registry + "/"
	}
	if repository != "" {
		imageName += repository + "/" + repoName
	} else {
		imageName += owner + "/" + repoName
	}
	// get branch from ref (if branch) or tag
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
	var buildReport BuildReport
	buildReport.BuildContext = options.RemoteContext
	buildReport.RepoName = repoName
	buildReport.Owner = owner
	buildReport.ImageName = imageName
	buildReport.Tag = tag
	buildReport.Start = time.Now()
	buildReport.StatusNotify = statusNotify

	// send build output and status
	if notify != nil {
		notify.SendBuildReport(ctx, buildResponse.Body, buildReport)
	}

	// push new image, if registry authBase64 is not nil (credentials specified)
	if api.authBase64 != "" {
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
			if notify != nil {
				notify.SendPushReport(ctx, pushResponse, image)
			}
		}
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
