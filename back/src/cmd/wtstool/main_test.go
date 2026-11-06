package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestHelpVersionAndCompletionWithoutConfig(t *testing.T) {
	t.Setenv("WTS_CONFIG", "")
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "root", args: []string{"wtstool"}, want: "wx"},
		{name: "help", args: []string{"wtstool", "--help"}, want: "jwt"},
		{name: "group help", args: []string{"wtstool", "wx", "--help"}, want: "WeChat"},
		{name: "leaf help", args: []string{"wtstool", "wx", "menu", "set", "--help"}, want: "group"},
		{name: "help with invalid config", args: []string{"wtstool", "-c", "/missing/config.yaml", "jwt", "issue", "--help"}, want: "expire-minutes"},
		{name: "version", args: []string{"wtstool", "--version"}, want: "dev"},
		{name: "completion", args: []string{"wtstool", "completion", "bash"}, want: "wtstool"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app := newApp()
			var output, errorOutput bytes.Buffer
			app.Writer = &output
			app.ErrWriter = &errorOutput
			if err := app.Run(context.Background(), test.args); err != nil {
				t.Fatalf("Run(%v): %v", test.args, err)
			}
			if !strings.Contains(output.String(), test.want) {
				t.Fatalf("output %q does not contain %q", output.String(), test.want)
			}
			if errorOutput.Len() != 0 {
				t.Fatalf("unexpected stderr: %s", errorOutput.String())
			}
		})
	}
}

func TestUnknownCommandAndLegacyActionFail(t *testing.T) {
	for _, args := range [][]string{
		{"wtstool", "unknown"},
		{"wtstool", "wx", "unknown"},
		{"wtstool", "-a", "get_wx_tags"},
	} {
		app := newApp()
		app.Writer = &bytes.Buffer{}
		app.ErrWriter = &bytes.Buffer{}
		if err := app.Run(context.Background(), args); err == nil {
			t.Errorf("Run(%v) unexpectedly succeeded", args)
		}
	}
}

func TestUsageErrorsDoNotWriteOutput(t *testing.T) {
	for _, args := range [][]string{
		{"-a", "get_wx_tags"},
		{"wx", "menu", "set"},
		{"wx", "tag", "list", "--unknown"},
		{"jwt", "issue", "--openid", "id", "--access", "user", "--expire-minutes", "invalid"},
	} {
		output, stderr, err := runTestApp(args...)
		if err == nil || output != "" || stderr != "" {
			t.Errorf("Run(%v): output=%q stderr=%q err=%v", args, output, stderr, err)
		}
	}
}
