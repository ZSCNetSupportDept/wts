package wechat

import (
	"net/url"
	"testing"
)

func TestBuildNotifyLanderURL(t *testing.T) {
	got, err := BuildNotifyLanderURL(
		"https://wwbx.davisye.cn/app/",
		42,
		"solved",
		"网线已更换 & 请测试网络",
	)
	if err != nil {
		t.Fatalf("BuildNotifyLanderURL() error = %v", err)
	}

	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("url.Parse(%q) error = %v", got, err)
	}
	if parsed.Scheme != "https" || parsed.Host != "wwbx.davisye.cn" {
		t.Fatalf("unexpected URL origin: %s://%s", parsed.Scheme, parsed.Host)
	}
	if parsed.Path != "/app/wx_notify_lander/" {
		t.Fatalf("path = %q, want %q", parsed.Path, "/app/wx_notify_lander/")
	}
	if gotTid := parsed.Query().Get("tid"); gotTid != "42" {
		t.Fatalf("tid = %q, want %q", gotTid, "42")
	}
	if gotStatus := parsed.Query().Get("status"); gotStatus != "solved" {
		t.Fatalf("status = %q, want %q", gotStatus, "solved")
	}
	if gotMessage := parsed.Query().Get("message"); gotMessage != "网线已更换 & 请测试网络" {
		t.Fatalf("message = %q", gotMessage)
	}
}

func TestBuildNotifyLanderURLRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		frontEndURL string
		tid         int32
		status      string
	}{
		{name: "missing URL", frontEndURL: "", tid: 1, status: "solved"},
		{name: "non HTTPS URL", frontEndURL: "http://example.test", tid: 1, status: "solved"},
		{name: "URL query", frontEndURL: "https://example.test/?foo=bar", tid: 1, status: "solved"},
		{name: "invalid ticket ID", frontEndURL: "https://example.test", tid: 0, status: "solved"},
		{name: "missing status", frontEndURL: "https://example.test", tid: 1, status: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := BuildNotifyLanderURL(test.frontEndURL, test.tid, test.status, "message"); err == nil {
				t.Fatal("BuildNotifyLanderURL() error = nil, want error")
			}
		})
	}
}
