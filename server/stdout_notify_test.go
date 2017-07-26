package main

import (
	"context"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

func TestStdoutNotify_SendBuildReport(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx    context.Context
		r      io.ReadCloser
		report BuildReport
	}
	tests := []struct {
		name string
		out  StdoutNotify
		args args
	}{
		{"empty", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader("")), BuildReport{}}},
		{"success", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader(`{"stream": "Successfully built abcd1234abcd"}`)), BuildReport{}}},
		{
			"success",
			StdoutNotify{},
			args{
				ctx,
				ioutil.NopCloser(strings.NewReader(`{"stream": "Successfully built abcd1234abcd"}`)),
				BuildReport{
					RepoName:     "test-repo",
					Owner:        "test-user",
					ImageName:    "test-repo/test-user",
					Tag:          "latest",
					BuildContext: "git://git/repository",
				},
			},
		},
		{"error", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader("ERROR")), BuildReport{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := StdoutNotify{}
			out.SendBuildReport(tt.args.ctx, tt.args.r, tt.args.report)
		})
	}
}

func TestStdoutNotify_SendPushReport(t *testing.T) {
	ctx := context.Background()
	type args struct {
		ctx   context.Context
		r     io.ReadCloser
		image string
	}
	tests := []struct {
		name string
		out  StdoutNotify
		args args
	}{
		{"oneline", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader(`{"stream": "output line"}`)), "test/image"}},
		{"multiline", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader(`{"stream": "output line"}
			{"stream": "second line"}
			{"stream": "third line"}`)), "test/image"}},
		{"empty", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader("")), "test/image"}},
		{"error", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader("ERROR")), "test/image"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := StdoutNotify{}
			out.SendPushReport(tt.args.ctx, tt.args.r, tt.args.image)
		})
	}
}
