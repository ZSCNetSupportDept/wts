package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/silenceper/wechat/v2/util"
)

func startWXServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/cgi-bin/token" {
			_, _ = io.WriteString(writer, `{"access_token":"sensitive-token","expires_in":7200}`)
			return
		}
		handler(writer, request)
	}))
	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	util.SetURIModifier(func(address string) string {
		parsed, err := url.Parse(address)
		if err != nil {
			return address
		}
		parsed.Scheme = baseURL.Scheme
		parsed.Host = baseURL.Host
		return parsed.String()
	})
	t.Cleanup(func() {
		util.SetURIModifier(nil)
		server.Close()
	})
}

func setWXConfig(t *testing.T) {
	t.Helper()
	t.Setenv("WTS_CONFIG", testConfig(t, "WX:\n  AppID: test-app\n  AppSecret: test-secret\nDebug:\n  ProgramVerbose: true\n"))
	t.Setenv("WTS_WX_APPID", "")
	t.Setenv("WTS_WX_APPSECRET", "")
}

func TestWXValidationAndUnimplementedMenu(t *testing.T) {
	t.Setenv("WTS_CONFIG", "")
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, []byte("{invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"wx", "menu", "get"}, want: "not implemented"},
		{args: []string{"wx", "menu", "set", "bad", path}, want: "unknown menu group"},
		{args: []string{"wx", "menu", "set", "default", path}, want: "invalid JSON"},
		{args: []string{"wx", "tag", "create", ""}, want: "tag name"},
		{args: []string{"wx", "tag", "list", "--output", "xml"}, want: "unknown output format"},
		{args: []string{"wx", "user", "tag", "id", "bad"}, want: "unknown tag"},
		{args: []string{"wx", "tag", "list"}, want: "configuration file"},
		{args: []string{"wx", "tag", "list", "extra"}, want: "unexpected argument"},
	}
	for _, test := range tests {
		_, _, err := runTestApp(test.args...)
		if err == nil || !strings.Contains(err.Error(), test.want) {
			t.Errorf("Run(%v) error = %v, want %q", test.args, err, test.want)
		}
	}
}

func TestWXMenuRequests(t *testing.T) {
	for _, test := range []struct {
		group string
		path  string
	}{
		{group: "default", path: "/cgi-bin/menu/create"},
		{group: "operator", path: "/cgi-bin/menu/addconditional"},
		{group: "admin", path: "/cgi-bin/menu/addconditional"},
	} {
		t.Run(test.group, func(t *testing.T) {
			setWXConfig(t)
			content := `{"button":[],"matchrule":{"group_id":"100"}}`
			path := filepath.Join(t.TempDir(), "menu with spaces.json")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			calls := 0
			startWXServer(t, func(writer http.ResponseWriter, request *http.Request) {
				calls++
				body, err := io.ReadAll(request.Body)
				if err != nil {
					t.Error(err)
				}
				if request.Method != http.MethodPost || request.URL.Path != test.path || string(body) != content {
					t.Errorf("unexpected request: %s %s %s", request.Method, request.URL.Path, body)
				}
				_, _ = io.WriteString(writer, `{"errcode":0,"errmsg":"ok"}`)
			})
			output, stderr, err := runTestApp("wx", "menu", "set", test.group, path)
			if err != nil || stderr != "" || !strings.Contains(output, "Menu updated") || calls != 1 {
				t.Fatalf("menu set: output=%q stderr=%q calls=%d err=%v", output, stderr, calls, err)
			}
		})
	}
}

func TestWXCreateTag(t *testing.T) {
	setWXConfig(t)
	calls := 0
	startWXServer(t, func(writer http.ResponseWriter, request *http.Request) {
		calls++
		var body struct {
			Tag struct {
				Name string `json:"name"`
			} `json:"tag"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if request.URL.Path != "/cgi-bin/tags/create" || body.Tag.Name != "tag with spaces" {
			t.Errorf("unexpected request: %s %+v", request.URL.Path, body)
		}
		_, _ = io.WriteString(writer, `{"tag":{"id":102,"name":"tag with spaces"}}`)
	})
	output, stderr, err := runTestApp("wx", "tag", "create", "tag with spaces")
	if err != nil || stderr != "" || !strings.Contains(output, "102: tag with spaces") || calls != 1 {
		t.Fatalf("tag create: output=%q stderr=%q calls=%d err=%v", output, stderr, calls, err)
	}
}

func TestWXListTags(t *testing.T) {
	tests := []struct {
		name     string
		response string
		status   int
		want     string
		wantErr  string
	}{
		{name: "tags", response: `{"tags":[{"id":100,"name":"team","count":3}]}`, want: `[{"id":100,"name":"team","count":3}]`},
		{name: "empty tags", response: `{"tags":[]}`, want: `[]`},
		{name: "null tags", response: `{"tags":null}`, want: `[]`},
		{name: "wechat error", response: `{"errcode":40001,"errmsg":"sensitive-token"}`, wantErr: "40001"},
		{name: "http error", status: http.StatusInternalServerError, wantErr: "request failed"},
		{name: "invalid JSON", response: `{invalid`, wantErr: "invalid WeChat response"},
		{name: "missing tags", response: `{}`, wantErr: "missing tags"},
		{name: "invalid tag", response: `{"tags":[null]}`, wantErr: "invalid WeChat tag"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setWXConfig(t)
			startWXServer(t, func(writer http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/cgi-bin/tags/get" || request.URL.Query().Get("access_token") != "sensitive-token" {
					t.Errorf("unexpected request path or access token")
				}
				if test.status != 0 {
					writer.WriteHeader(test.status)
				}
				_, _ = io.WriteString(writer, test.response)
			})
			output, stderr, err := runTestApp("wx", "tag", "list", "--output", "json")
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("error = %v, want %q", err, test.wantErr)
				}
				for _, secret := range []string{"test-secret", "sensitive-token", "access_token="} {
					if strings.Contains(err.Error(), secret) || strings.Contains(stderr, secret) || strings.Contains(output, secret) {
						t.Errorf("error output leaked %q", secret)
					}
				}
				return
			}
			if err != nil || stderr != "" || strings.TrimSpace(output) != test.want {
				t.Fatalf("list tags: output=%q stderr=%q err=%v, want %q", output, stderr, err, test.want)
			}
		})
	}
}

func TestWXTokenErrorDoesNotExposeCredentials(t *testing.T) {
	setWXConfig(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	}))
	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	util.SetURIModifier(func(address string) string {
		parsed, err := url.Parse(address)
		if err != nil {
			return address
		}
		parsed.Scheme = baseURL.Scheme
		parsed.Host = baseURL.Host
		return parsed.String()
	})
	t.Cleanup(func() {
		util.SetURIModifier(nil)
		server.Close()
	})
	output, stderr, err := runTestApp("wx", "tag", "list", "-o", "json")
	if err == nil {
		t.Fatal("expected token request to fail")
	}
	for _, secret := range []string{"test-secret", "secret=", "sensitive-token"} {
		if strings.Contains(err.Error(), secret) || strings.Contains(output, secret) || strings.Contains(stderr, secret) {
			t.Errorf("output leaked %q", secret)
		}
	}
}

func TestWXChangeUserTag(t *testing.T) {
	type tagRequest struct {
		path string
		id   int32
	}
	tests := []struct {
		name       string
		tag        string
		firstCode  int
		secondCode int
		want       []tagRequest
		wantErr    string
	}{
		{name: "default", tag: "default", want: []tagRequest{{"batchuntagging", 100}, {"batchuntagging", 101}}},
		{name: "operator", tag: "operator", firstCode: 45059, want: []tagRequest{{"batchuntagging", 101}, {"batchtagging", 100}}},
		{name: "admin", tag: "admin", want: []tagRequest{{"batchuntagging", 100}, {"batchtagging", 101}}},
		{name: "first failure", tag: "operator", firstCode: 40003, want: []tagRequest{{"batchuntagging", 101}}, wantErr: "40003"},
		{name: "second failure", tag: "admin", secondCode: 40004, want: []tagRequest{{"batchuntagging", 100}, {"batchtagging", 101}}, wantErr: "40004"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setWXConfig(t)
			var requests []tagRequest
			startWXServer(t, func(writer http.ResponseWriter, request *http.Request) {
				var body struct {
					OpenIDs []string `json:"openid_list"`
					TagID   int32    `json:"tagid"`
				}
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(body.OpenIDs, []string{"openid"}) {
					t.Errorf("unexpected OpenIDs: %v", body.OpenIDs)
				}
				requests = append(requests, tagRequest{filepath.Base(request.URL.Path), body.TagID})
				code := test.secondCode
				if len(requests) == 1 {
					code = test.firstCode
				}
				_, _ = fmt.Fprintf(writer, `{"errcode":%d,"errmsg":"sensitive-token"}`, code)
			})
			_, _, err := runTestApp("wx", "user", "tag", "openid", test.tag)
			if test.wantErr == "" && err != nil {
				t.Fatalf("change user tag: %v", err)
			}
			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr) || strings.Contains(err.Error(), "sensitive-token")) {
				t.Fatalf("error = %v, want code %s without token", err, test.wantErr)
			}
			if !reflect.DeepEqual(requests, test.want) {
				t.Errorf("requests = %v, want %v", requests, test.want)
			}
		})
	}
}
