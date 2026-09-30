package wechat

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// BuildNotifyLanderURL builds the H5 page opened from a WeChat status notification.
func BuildNotifyLanderURL(frontEndURL string, tid int32, status string, message string) (string, error) {
	if tid <= 0 {
		return "", fmt.Errorf("invalid ticket ID: %d", tid)
	}
	if status == "" {
		return "", fmt.Errorf("notification status is required")
	}

	base, err := url.Parse(frontEndURL)
	if err != nil {
		return "", fmt.Errorf("parse frontend URL: %w", err)
	}
	if base.Scheme != "https" || base.Host == "" {
		return "", fmt.Errorf("frontend URL must be an absolute HTTPS URL")
	}
	if base.RawQuery != "" || base.Fragment != "" {
		return "", fmt.Errorf("frontend URL must not contain a query or fragment")
	}

	base.Path = strings.TrimRight(base.Path, "/") + "/wx_notify_lander/"
	base.RawPath = ""
	base.RawQuery = url.Values{
		"tid":     {strconv.FormatInt(int64(tid), 10)},
		"status":  {status},
		"message": {message},
	}.Encode()
	return base.String(), nil
}
