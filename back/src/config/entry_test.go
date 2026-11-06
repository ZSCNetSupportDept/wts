package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func writeTestConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFileDefaultsAndEnvironment(t *testing.T) {
	path := writeTestConfig(t, "Brand: file brand\nWX:\n  AppID: file-app\n")
	t.Setenv("WTS_LISTENPORT", "")
	t.Setenv("WTS_DB_TYPE", "")
	t.Setenv("WTS_FRONTEND_ONAUTHSUCCESS", "")
	t.Setenv("WTS_WX_APPID", "env-app")
	t.Setenv("WTS_WX_APPSECRET", "env-secret")
	t.Setenv("WTS_JWTKEY", "env-key")

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Brand != "file brand" || cfg.WX.AppID != "env-app" || cfg.WX.AppSecret != "env-secret" || cfg.JWTKey != "env-key" {
		t.Fatalf("unexpected config values: brand=%q appid=%q appsecret=%q jwtkey=%q", cfg.Brand, cfg.WX.AppID, cfg.WX.AppSecret, cfg.JWTKey)
	}
	if cfg.ListenPort != 25005 || cfg.DB.Type != "PostgreSQL" || cfg.FrontEnd.OnAuthSuccess != "/auth_success.html" {
		t.Fatalf("unexpected defaults: port=%d db=%q redirect=%q", cfg.ListenPort, cfg.DB.Type, cfg.FrontEnd.OnAuthSuccess)
	}
}

func TestLoadServerFlagPrecedence(t *testing.T) {
	path := writeTestConfig(t, "ListenPort: 10001\nDB:\n  User: file-user\n")
	t.Setenv("WTS_LISTENPORT", "10002")
	t.Setenv("WTS_DB_USER", "env-user")
	flags := pflag.NewFlagSet("server", pflag.ContinueOnError)
	registerFlags(flags)
	if err := flags.Parse([]string{"--ListenPort", "10003"}); err != nil {
		t.Fatal(err)
	}

	cfg, err := load(path, flags)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenPort != 10003 || cfg.DB.User != "env-user" {
		t.Fatalf("unexpected precedence: port=%d user=%q", cfg.ListenPort, cfg.DB.User)
	}
}

func TestLoadFileErrors(t *testing.T) {
	invalid := writeTestConfig(t, "WX: [invalid\n")
	for _, path := range []string{"", filepath.Join(t.TempDir(), "missing.yaml"), invalid} {
		if _, err := LoadFile(path); err == nil {
			t.Errorf("LoadFile(%q) unexpectedly succeeded", path)
		}
	}
}

func TestLoadServerWrapper(t *testing.T) {
	path := writeTestConfig(t, "ListenPort: 10001\nDebug:\n  ProgramVerbose: false\n")
	tests := []struct {
		name    string
		mode    string
		want    string
		wantErr bool
	}{
		{name: "flag overrides env and file", mode: "config", want: "port=10003"},
		{name: "missing config", mode: "missing", want: "config file path must be provided", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := exec.Command(os.Args[0], "-test.run=^TestLoadServerHelperProcess$")
			command.Env = append(os.Environ(),
				"WTS_TEST_CONFIG_MODE="+test.mode,
				"WTS_TEST_CONFIG_PATH="+path,
				"WTS_LISTENPORT=10002",
				"WTS_DEBUG_PROGRAMVERBOSE=false",
			)
			output, err := command.CombinedOutput()
			if (err != nil) != test.wantErr || !strings.Contains(string(output), test.want) {
				t.Fatalf("server config: output=%q err=%v", output, err)
			}
		})
	}
}

func TestLoadServerHelperProcess(t *testing.T) {
	mode := os.Getenv("WTS_TEST_CONFIG_MODE")
	if mode == "" {
		return
	}
	pflag.CommandLine = pflag.NewFlagSet("wts-server", pflag.ExitOnError)
	os.Args = []string{"wts-server"}
	if mode == "config" {
		os.Args = append(os.Args, "-c", os.Getenv("WTS_TEST_CONFIG_PATH"), "--ListenPort", "10003")
	}
	cfg := Load()
	fmt.Printf("port=%d\n", cfg.ListenPort)
	os.Exit(0)
}
