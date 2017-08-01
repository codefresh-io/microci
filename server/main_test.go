package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

type DockerMock struct {
	mock.Mock
}

func (m *DockerMock) Info(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *DockerMock) RegistryLogin(ctx context.Context, user, password, registry string) error {
	args := m.Called(ctx, user, password, registry)
	return args.Error(0)
}
func (m *DockerMock) BuildPushImage(ctx context.Context, cloneURL, ref, name, fullname, tag, registry, repository string, notify BuildNotify, statusNotify GitStatusNotify) error {
	args := m.Called(ctx, cloneURL, ref, name, fullname, tag, registry, repository, notify)
	return args.Error(0)
}

var dockerMock *DockerMock
var getMockDockerClient = func() DockerClient {
	if dockerMock == nil {
		dockerMock = &DockerMock{}
	}
	return dockerMock
}

//---- TESTS

func Test_main(t *testing.T) {
	os.Args = []string{"microci", "-v"}
	main()
}

func Test_before(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	globalSet := flag.NewFlagSet("test", 0)
	globalCtx := cli.NewContext(nil, globalSet, nil)

	c := cli.NewContext(nil, set, globalCtx)

	globalSet.Bool("debug", true, "doc")
	c2 := cli.NewContext(nil, set, globalCtx)

	globalSet.Bool("json", true, "doc")
	c3 := cli.NewContext(nil, set, globalCtx)

	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"No Error", args{c}, false},
		{"Debug", args{c2}, false},
		{"JSON", args{c3}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := before(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("before() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_handleSignals(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())
	gCancelCommands.Append(cancel)
	type args struct {
		sigs         chan os.Signal
		exitOnSignal bool
	}
	tests := []struct {
		name string
		args args
		sig  os.Signal
	}{
		{"SIGTERM", args{make(chan os.Signal, 1), false}, syscall.SIGTERM},
		{"SIGKILL", args{make(chan os.Signal, 1), false}, syscall.SIGKILL},
		{"UNKNOWN", args{make(chan os.Signal, 1), false}, syscall.SIGUSR1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleSignals(tt.args.sigs, tt.args.exitOnSignal)
			tt.args.sigs <- tt.sig
			// wait a while to handle signal
			time.Sleep(time.Millisecond)
		})
	}
}

func Test_handleWebhook(t *testing.T) {
	// simple set of flags
	set := flag.NewFlagSet("test", 0)
	set.String("secret", "test-secret", "doc")
	set.String("registry", "test-registry", "doc")
	set.String("repository", "test-repository", "doc")
	// same as above, plus credentials
	setLogin := flag.NewFlagSet("test", 0)
	setLogin.String("secret", "test-secret", "doc")
	setLogin.String("user", "test-user", "doc")
	setLogin.String("password", "test-password", "doc")
	setLogin.String("registry", "test-registry", "doc")
	setLogin.String("repository", "test-repository", "doc")

	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name string
		args args
	}{
		{"testNoLogin", args{cli.NewContext(nil, set, nil)}},
		{"testRegistryLogin", args{cli.NewContext(nil, setLogin, nil)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handleWebhook(tt.args.c)
			server := httptest.NewServer(h)
			defer server.Close()
			// invoke method
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("Received non-200 response: %d\n", resp.StatusCode)
			}
		})
	}
}

func Test_statusHandler(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{"noDebug", args{httptest.NewRecorder(), httptest.NewRequest("GET", "http://test/status", nil)}, ASCIILogo},
		{"withDebug", args{httptest.NewRecorder(), httptest.NewRequest("GET", "http://test/status?debug=true", nil)}, "Debug request:"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusHandler(tt.args.w, tt.args.r)
			resp := tt.args.w.Result()
			body, _ := ioutil.ReadAll(resp.Body)
			assert.True(t, strings.Contains(string(body), tt.expected))
		})
	}
}

func Test_dockerInfo(t *testing.T) {
	keepGetDockerClient := getDockerClient
	defer func() { getDockerClient = keepGetDockerClient }()
	getDockerClient = getMockDockerClient
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name       string
		args       args
		mockClient *DockerMock
		err        error
	}{
		{"infoNoError", args{nil}, getDockerClient().(*DockerMock), nil},
		{"infoWithError", args{nil}, getDockerClient().(*DockerMock), errors.New("Test Error")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClient.On("Info", mock.Anything).Return("Test Info", tt.err)
			dockerInfo(tt.args.c)
			tt.mockClient.AssertExpectations(t)
		})
	}
}
