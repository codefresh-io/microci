package main

import (
	"flag"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/urfave/cli"
)

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
	sigs := make(chan os.Signal, 1)
	type args struct {
		sigs         chan os.Signal
		exitOnSignal bool
	}
	tests := []struct {
		name string
		args args
	}{
		{"SIGTERM", args{sigs, false}},
		{"SIGKILL", args{sigs, false}},
		{"UNKNOWN", args{sigs, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleSignals(tt.args.sigs, tt.args.exitOnSignal)
			if tt.name == "SIGTERM" {
				// send SIGTERM signal
				sigs <- syscall.SIGTERM
			} else if tt.name == "SIGKILL" {
				// send SIGKILL signal
				sigs <- syscall.SIGKILL
			} else {
				sigs <- syscall.SIGUSR1
			}
			// wait a while to handle signal
			time.Sleep(time.Millisecond)
		})
	}
}

func Test_beforeCommand(t *testing.T) {
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := beforeCommand(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("beforeCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_webhookServer(t *testing.T) {
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webhookServer(tt.args.c)
		})
	}
}

func Test_reportHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reportHandler(tt.args.w, tt.args.r)
		})
	}
}

func Test_statusHandler(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusHandler(tt.args.w, tt.args.r)
		})
	}
}

func Test_dockerInfo(t *testing.T) {
	type args struct {
		c *cli.Context
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerInfo(tt.args.c)
		})
	}
}
