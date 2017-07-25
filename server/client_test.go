package main

import (
	"errors"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

type DockerClientAPIMock struct {
	mock.Mock
}

func (m *DockerClientAPIMock) Info(ctx context.Context) (types.Info, error) {
	args := m.Called(ctx)
	return args.Get(0).(types.Info), args.Error(1)
}

func (m *DockerClientAPIMock) RegistryLogin(ctx context.Context, auth types.AuthConfig) (registry.AuthenticateOKBody, error) {
	args := m.Called(ctx, auth)
	return args.Get(0).(registry.AuthenticateOKBody), args.Error(1)
}

func (m *DockerClientAPIMock) ImagePush(ctx context.Context, ref string, options types.ImagePushOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *DockerClientAPIMock) ImageBuild(ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	args := m.Called(ctx, buildContext, options)
	return args.Get(0).(types.ImageBuildResponse), args.Error(1)
}

func Test_dockerAPI_RegistryLogin(t *testing.T) {
	type fields struct {
		apiClient  DockerClientAPI
		authBase64 string
	}
	type args struct {
		ctx      context.Context
		user     string
		password string
		registry string
	}
	type results struct {
		authOKBody registry.AuthenticateOKBody
		err        error
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		results results
	}{
		// Valid Login test case
		{
			"validLogin",
			fields{
				apiClient: &DockerClientAPIMock{},
				// base64 encoded JSON with user and password
				authBase64: "ewpVc2VybmFtZTogdGVzdC11c2VyLApQYXNzd29yZDogdGVzdC1wYXNzd29yZAp9",
			},
			args{
				ctx:      context.Background(),
				user:     "test-user",
				password: "test-password",
				registry: "test-registry",
			},
			false,
			results{
				authOKBody: registry.AuthenticateOKBody{IdentityToken: "good-token", Status: "OK"},
				err:        nil,
			},
		},
		// Inalid Login test case
		{
			"invalidLogin",
			fields{
				apiClient:  &DockerClientAPIMock{},
				authBase64: "XXX",
			},
			args{
				ctx:      context.Background(),
				user:     "test-user",
				password: "test-password",
				registry: "test-registry",
			},
			true,
			results{
				authOKBody: registry.AuthenticateOKBody{},
				err:        errors.New("Unauthorized"),
			},
		},
		// empty credentials test case
		{
			"emptyCredentials",
			fields{
				apiClient:  &DockerClientAPIMock{},
				authBase64: "XXX",
			},
			args{
				ctx:      context.Background(),
				user:     "",
				password: "",
				registry: "test-registry",
			},
			false,
			results{
				authOKBody: registry.AuthenticateOKBody{},
				err:        nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &dockerAPI{
				apiClient:  tt.fields.apiClient,
				authBase64: tt.fields.authBase64,
			}
			authConfig := types.AuthConfig{Username: tt.args.user, Password: tt.args.password, ServerAddress: tt.args.registry}
			if tt.args.user != "" && tt.args.password != "" {
				api.apiClient.(*DockerClientAPIMock).On("RegistryLogin", tt.args.ctx, authConfig).Return(tt.results.authOKBody, tt.results.err)
			}
			if err := api.RegistryLogin(tt.args.ctx, tt.args.user, tt.args.password, tt.args.registry); (err != nil) != tt.wantErr {
				t.Errorf("dockerAPI.RegistryLogin() error = %v, wantErr %v", err, tt.wantErr)
			}
			api.apiClient.(*DockerClientAPIMock).AssertExpectations(t)
			if tt.args.user == "" && tt.args.password == "" {
				api.apiClient.(*DockerClientAPIMock).AssertNumberOfCalls(t, "RegistryLogin", 0)
			}
		})
	}
}

func Test_dockerAPI_BuildPushImage(t *testing.T) {
	type fields struct {
		apiClient  DockerClientAPI
		authBase64 string
	}
	type args struct {
		ctx        context.Context
		cloneURL   string
		ref        string
		name       string
		fullname   string
		tag        string
		registry   string
		repository string
		notify     BuildNotify
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &dockerAPI{
				apiClient:  tt.fields.apiClient,
				authBase64: tt.fields.authBase64,
			}
			if err := api.BuildPushImage(tt.args.ctx, tt.args.cloneURL, tt.args.ref, tt.args.name, tt.args.fullname, tt.args.tag, tt.args.registry, tt.args.repository, tt.args.notify); (err != nil) != tt.wantErr {
				t.Errorf("dockerAPI.BuildPushImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dockerAPI_Info(t *testing.T) {
	type fields struct {
		apiClient  DockerClientAPI
		authBase64 string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &dockerAPI{
				apiClient:  tt.fields.apiClient,
				authBase64: tt.fields.authBase64,
			}
			got, err := api.Info(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("dockerAPI.Info() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("dockerAPI.Info() = %v, want %v", got, tt.want)
			}
		})
	}
}
