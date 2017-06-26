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
		target BuildTarget
	}
	tests := []struct {
		name string
		out  StdoutNotify
		args args
	}{
		{"empty", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader("")), BuildTarget{}}},
		{"success", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader(`{"stream": "Successfully built abcd1234abcd"}`)), BuildTarget{}}},
		{"success", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader(`{"stream": "Successfully built abcd1234abcd"}`)), BuildTarget{"test/image", "latest", "git://git/repository"}}},
		{"error", StdoutNotify{}, args{ctx, ioutil.NopCloser(strings.NewReader("ERROR")), BuildTarget{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := StdoutNotify{}
			out.SendBuildReport(tt.args.ctx, tt.args.r, tt.args.target)
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
