package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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

func TestParseJWT(t *testing.T) {
	t.Setenv("WTS_JWTKEY", "")
	t.Setenv("WTS_CONFIG", testConfig(t, "JWTKey: signing-key\n"))
	tests := []struct {
		name      string
		key       string
		expires   time.Time
		notBefore time.Time
		verified  bool
	}{
		{name: "valid", key: "signing-key", expires: time.Now().Add(time.Hour), verified: true},
		{name: "expired", key: "signing-key", expires: time.Now().Add(-time.Hour), verified: true},
		{name: "not yet valid", key: "signing-key", expires: time.Now().Add(2 * time.Hour), notBefore: time.Now().Add(time.Hour), verified: true},
		{name: "wrong key", key: "other-key", expires: time.Now().Add(time.Hour)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			claims := jwt.MapClaims{
				"sub":       "example",
				"exp":       test.expires.Unix(),
				"large_num": json.Number("9007199254740993"),
			}
			if !test.notBefore.IsZero() {
				claims["nbf"] = test.notBefore.Unix()
			}
			signed := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signed.Header["kid"] = "test-key"
			token, err := signed.SignedString([]byte(test.key))
			if err != nil {
				t.Fatal(err)
			}
			output, stderr, err := runTestApp("jwt", "parse", token)
			if (err == nil) != test.verified || stderr != "" {
				t.Fatalf("parse JWT: err=%v stderr=%q", err, stderr)
			}
			var result struct {
				Header   map[string]any `json:"header"`
				Payload  map[string]any `json:"payload"`
				Verified bool           `json:"verified"`
			}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("invalid JSON output %q: %v", output, err)
			}
			if result.Header["alg"] != "HS256" || result.Header["kid"] != "test-key" || result.Payload["sub"] != "example" || result.Verified != test.verified {
				t.Fatalf("unexpected result: %+v", result)
			}
			if !strings.Contains(output, "9007199254740993") {
				t.Fatalf("large number lost precision in output: %s", output)
			}
		})
	}
}

func TestParseJWTRejectsUntrustedSignatures(t *testing.T) {
	t.Setenv("WTS_JWTKEY", "")
	t.Setenv("WTS_CONFIG", testConfig(t, "JWTKey: signing-key\n"))
	claims := jwt.MapClaims{"sub": "original"}
	valid, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("signing-key"))
	if err != nil {
		t.Fatal(err)
	}
	hs384, err := jwt.NewWithClaims(jwt.SigningMethodHS384, claims).SignedString([]byte("signing-key"))
	if err != nil {
		t.Fatal(err)
	}
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(valid, ".")
	tamperedPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"tampered"}`))
	unknownHeader := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"unknown"}`))
	tests := []struct {
		name    string
		token   string
		subject string
	}{
		{name: "tampered payload", token: strings.Join([]string{parts[0], tamperedPayload, parts[2]}, "."), subject: "tampered"},
		{name: "different HMAC algorithm", token: hs384, subject: "original"},
		{name: "unsigned", token: unsigned, subject: "original"},
		{name: "unknown algorithm", token: strings.Join([]string{unknownHeader, parts[1], parts[2]}, "."), subject: "original"},
		{name: "invalid signature encoding", token: strings.Join([]string{parts[0], parts[1], "!"}, "."), subject: "original"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, stderr, err := runTestApp("jwt", "parse", test.token)
			if err == nil || stderr != "" {
				t.Fatalf("parse JWT: err=%v stderr=%q", err, stderr)
			}
			var result struct {
				Payload  map[string]any `json:"payload"`
				Verified bool           `json:"verified"`
			}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("invalid JSON output %q: %v", output, err)
			}
			if result.Verified || result.Payload["sub"] != test.subject {
				t.Fatalf("unexpected result: %+v", result)
			}
		})
	}
}

func TestParseJWTRejectsInvalidInput(t *testing.T) {
	t.Setenv("WTS_CONFIG", "")
	t.Setenv("WTS_JWTKEY", "")
	path := testConfig(t, "JWTKey: signing-key\n")
	emptyKeyPath := testConfig(t, "JWTKey: ''\n")
	validJSON := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
	invalidJSON := base64.RawURLEncoding.EncodeToString([]byte(`not json`))
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing token", args: []string{"jwt", "parse"}, want: "token"},
		{name: "extra argument", args: []string{"-c", path, "jwt", "parse", "first", "second"}, want: "unexpected argument"},
		{name: "missing config", args: []string{"jwt", "parse", "a.b.c"}, want: "configuration file"},
		{name: "empty key", args: []string{"-c", emptyKeyPath, "jwt", "parse", "a.b.c"}, want: "JWTKey"},
		{name: "bad segment count", args: []string{"-c", path, "jwt", "parse", "a.b"}, want: "three segments"},
		{name: "bad header encoding", args: []string{"-c", path, "jwt", "parse", "!." + validJSON + ".sig"}, want: "decode JWT header"},
		{name: "bad header JSON", args: []string{"-c", path, "jwt", "parse", invalidJSON + "." + validJSON + ".sig"}, want: "decode JWT header"},
		{name: "bad payload encoding", args: []string{"-c", path, "jwt", "parse", validJSON + ".!.sig"}, want: "decode JWT payload"},
		{name: "bad payload JSON", args: []string{"-c", path, "jwt", "parse", validJSON + "." + invalidJSON + ".sig"}, want: "decode JWT payload"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, stderr, err := runTestApp(test.args...)
			if err == nil || !strings.Contains(err.Error(), test.want) || output != "" || stderr != "" {
				t.Fatalf("parse JWT: output=%q stderr=%q err=%v", output, stderr, err)
			}
		})
	}
}

func TestParseIssuedJWTWithEnvironmentKey(t *testing.T) {
	previous := hutil.NewWtsJWT
	t.Cleanup(func() { hutil.NewWtsJWT = previous })
	t.Setenv("WTS_CONFIG", testConfig(t, "JWTKey: wrong-key\n"))
	t.Setenv("WTS_JWTKEY", "environment-key")
	token, stderr, err := runTestApp("jwt", "issue", "--openid", "openid", "--access", "user", "--expire-minutes", "30")
	if err != nil || stderr != "" {
		t.Fatalf("issue JWT: err=%v stderr=%q", err, stderr)
	}
	output, stderr, err := runTestApp("jwt", "parse", strings.TrimSpace(token))
	if err != nil || stderr != "" {
		t.Fatalf("parse JWT: err=%v stderr=%q", err, stderr)
	}
	var result struct {
		Payload  map[string]any `json:"payload"`
		Verified bool           `json:"verified"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output %q: %v", output, err)
	}
	if !result.Verified || result.Payload["openid"] != "openid" || result.Payload["access"] != "user" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
