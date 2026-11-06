package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
)

func testConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runTestApp(args ...string) (string, string, error) {
	app := newApp()
	var output, errorOutput bytes.Buffer
	app.Writer = &output
	app.ErrWriter = &errorOutput
	err := app.Run(context.Background(), append([]string{"wtstool"}, args...))
	return output.String(), errorOutput.String(), err
}

func TestIssueJWT(t *testing.T) {
	previous := hutil.NewWtsJWT
	t.Cleanup(func() { hutil.NewWtsJWT = previous })
	t.Setenv("WTS_JWTKEY", "")
	t.Setenv("WTS_CONFIG", testConfig(t, "JWTKey: wrong-key\n"))
	path := testConfig(t, "JWTKey: signing-key\nDebug:\n  ProgramVerbose: true\n")
	tests := []struct {
		name    string
		expire  string
		args    []string
		minutes int64
	}{
		{name: "config before command", expire: "30", args: []string{"-c", path}, minutes: 30},
		{name: "config after command", expire: "0", args: []string{"--config", path}, minutes: 0},
		{name: "negative expiry", expire: "-5", args: []string{"--config", path}, minutes: -5},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := []string{"jwt", "issue", "--openid", "openid", "--access", "user", "--expire-minutes", test.expire,
				"--sid", "123", "--username", "A Name", "--avatar", "https://example.test/a", "--name", "Name"}
			if test.name == "config before command" {
				args = append(test.args, args...)
			} else {
				args = append(args, test.args...)
			}
			output, stderr, err := runTestApp(args...)
			if err != nil || stderr != "" {
				t.Fatalf("issue JWT: err=%v stderr=%q", err, stderr)
			}
			claims := &hutil.WtsJWT{}
			token, err := jwt.ParseWithClaims(strings.TrimSpace(output), claims, func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					t.Errorf("unexpected signing method %v", token.Method)
				}
				return []byte("signing-key"), nil
			}, jwt.WithoutClaimsValidation())
			if err != nil || !token.Valid {
				t.Fatalf("invalid JWT: %v", err)
			}
			if claims.OpenID != "openid" || claims.Sid != "123" || claims.Username != "A Name" || claims.Avatar != "https://example.test/a" || claims.Name != "Name" || claims.Access != "user" {
				t.Fatalf("unexpected claims: %+v", claims)
			}
			if got := int64(claims.ExpiresAt.Sub(claims.IssuedAt.Time) / time.Minute); got != test.minutes {
				t.Fatalf("expiry = %d minutes, want %d", got, test.minutes)
			}
		})
	}
}

func TestIssueJWTRejectsInvalidInput(t *testing.T) {
	t.Setenv("WTS_CONFIG", "")
	t.Setenv("WTS_JWTKEY", "")
	path := testConfig(t, "JWTKey: ''\n")
	tests := [][]string{
		{"jwt", "issue", "--openid", "id", "--access", "user", "--expire-minutes", "1"},
		{"-c", path, "jwt", "issue", "--openid", "id", "--access", "user", "--expire-minutes", "1"},
		{"jwt", "issue", "--openid", "id", "--access", "invalid", "--expire-minutes", "1"},
		{"jwt", "issue", "--openid", "id", "--access", "user"},
		{"jwt", "issue", "--openid", "id", "--access", "user", "--expire-minutes", "oops"},
		{"jwt", "issue", "--openid", "id", "--access", "user", "--expire-minutes", "1", "extra"},
	}
	for _, args := range tests {
		if _, _, err := runTestApp(args...); err == nil {
			t.Errorf("Run(%v) unexpectedly succeeded", args)
		}
	}
}

func TestIssueJWTWithEnvironmentAndAllRoles(t *testing.T) {
	previous := hutil.NewWtsJWT
	t.Cleanup(func() { hutil.NewWtsJWT = previous })
	t.Setenv("WTS_CONFIG", testConfig(t, "Brand: test\n"))
	t.Setenv("WTS_JWTKEY", "environment-key")
	for _, access := range []string{
		"dev", "chief", "api", "user", "unregistered", "formal-member",
		"informal-member", "pre-member", "group-leader",
	} {
		t.Run(access, func(t *testing.T) {
			output, stderr, err := runTestApp("jwt", "issue", "--openid", "", "--access", access, "--expire-minutes", "-1")
			if err != nil || stderr != "" {
				t.Fatalf("issue JWT: err=%v stderr=%q", err, stderr)
			}
			claims := &hutil.WtsJWT{}
			token, err := jwt.ParseWithClaims(strings.TrimSpace(output), claims, func(token *jwt.Token) (interface{}, error) {
				return []byte("environment-key"), nil
			}, jwt.WithoutClaimsValidation())
			if err != nil || !token.Valid || string(claims.Access) != access || claims.OpenID != "" {
				t.Fatalf("invalid JWT: claims=%+v err=%v", claims, err)
			}
		})
	}
}
