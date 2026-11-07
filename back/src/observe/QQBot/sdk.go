package qqbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"zsxyww.com/wts/server"
)

// 全局HTTP客户端，用于发送QQBot SDK的HTTP请求
var c *http.Client

// 全局 accessToken，用于存储QQBot SDK的访问令牌
var accessToken string
var tokenMu sync.RWMutex

// initClient initializes the HTTP client if it hasn't been initialized yet.
var initClient = func() {
	if c == nil {
		slog.Info("初始化QQBot SDK HTTP客户端")
		c = &http.Client{
			Timeout:   30 * time.Second,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		}
		refreshTokenOnce(server.Cfg.QQBot.AppID, server.Cfg.QQBot.AppSecret)
		go refreshToken()
	}
}

var initRequest = func(method, url string) (*http.Request, context.Context, QQBotError) {
	initClient()
	ctx := context.Background()
	req, _ := http.NewRequestWithContext(ctx, method, url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "wts QQBot SDK")
	tokenMu.RLock()
	req.Header.Set("Authorization", "QQBot "+accessToken)
	tokenMu.RUnlock()

	var e QQBotError // 初始化每个请求的错误对象

	slog.Debug("初始化QQBot SDK HTTP请求", "method", method, "url", url)
	return req, ctx, e
}

type QQBotError struct {
	// scope是本SDK的自定义标识，用于区分错误来源
	// Scope = "client" 表示错误来源于客户端，例如网络请求失败等
	// Scope = "api" 表示错误来源于QQBot API，例如返回400,500等HTTP错误码
	// Scope = "business" 表示错误来源于业务逻辑，QQ对业务逻辑错误会返回200，这里分装一下
	Scope   string      `json:"scope"`
	Code    json.Number `json:"code"`
	Message string      `json:"message"`
}

func (e QQBotError) Error() string {
	switch e.Scope {
	case "client":
		return fmt.Sprintf("请求错误: %s", e.Message)
	case "api":
		return fmt.Sprintf("QQBot API错误: %s (HTTP %s)", e.Message, e.Code)
	case "business":
		return fmt.Sprintf("业务逻辑错误: %s (%s)", e.Message, e.Code)
	default:
		return fmt.Sprintf("未知错误: %s", e.Message)
	}
}

// https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/access-token.html#%E8%8E%B7%E5%8F%96-access-token
func refreshToken() {
	sleepTime := 1 * time.Second
	for {
		time.Sleep(sleepTime)
		sleepTime = refreshTokenOnce(server.Cfg.QQBot.AppID, server.Cfg.QQBot.AppSecret)
	}
}

func refreshTokenOnce(id, secret string) time.Duration {
	res, err := GetAccessToken(GetAccessTokenReq{
		AppID:        id,
		ClientSecret: secret,
	})
	if err != nil {
		slog.Error("刷新QQBot SDK accessToken失败", "error", err)
		return 5 * time.Second
	}
	tokenMu.Lock()
	accessToken = res.AccessToken
	tokenMu.Unlock()
	expire, err := res.ExpiresIn.Int64()
	if err != nil {
		slog.Error("解析QQBot SDK accessToken过期时间失败", "error", err)
		return 5 * time.Second
	}
	expiresIn := time.Duration(expire)*time.Second - 55*time.Second
		return expiresIn
}

// 官方文档参考：https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/getting-started.html
const QQBOT_API_BASE = "https://api.bot.qq.com"

// https://api.bot.qq.com/app/getAppAccessToken
const API_GET_ACCESS_TOKEN = QQBOT_API_BASE + "/app/getAppAccessToken"

type GetAccessTokenReq struct {
	AppID        string `json:"appId"`
	ClientSecret string `json:"clientSecret"`
}

type GetAccessTokenRes struct {
	AccessToken string      `json:"access_token"`
	ExpiresIn   json.Number `json:"expires_in"`
}

type GetAccessTokenResFailed struct {
	Code    json.Number `json:"code"`
	Message string      `json:"message"`
}

func (e GetAccessTokenResFailed) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func GetAccessToken(req GetAccessTokenReq) (GetAccessTokenRes, error) {
	ctx := context.Background()
	r, _ := http.NewRequestWithContext(ctx, http.MethodPost, API_GET_ACCESS_TOKEN, nil)
	r.Header.Set("Accept", "application/json")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("User-Agent", "wts QQBot SDK")
	var e QQBotError
	_ = ctx

	//构造请求体
	body, err := json.Marshal(req)
	if err != nil {
		e.Scope = "client"
		e.Message = "序列化请求体失败：" + err.Error()
		return GetAccessTokenRes{}, e
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		e.Scope = "client"
		e.Message = "执行HTTP请求失败: " + err.Error()
		return GetAccessTokenRes{}, e
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		e.Scope = "api"
		e.Code = json.Number(fmt.Sprintf("%d", res.StatusCode))
		bodyBytes, _ := io.ReadAll(res.Body)
		e.Message = "QQBot API返回错误: " + e.Code.String() + ": " + APIStatusCode(res.StatusCode) + " " + fmt.Sprintf("%s", bodyBytes)
		return GetAccessTokenRes{}, e
	}

	var successRes GetAccessTokenRes
	var failedRes GetAccessTokenResFailed

	bodyBytes, _ := io.ReadAll(res.Body)
	if err := json.Unmarshal(bodyBytes, &successRes); err == nil && successRes.AccessToken != "" {
		return successRes, nil
	}
	if err := json.Unmarshal(bodyBytes, &failedRes); err == nil && failedRes.Code != "" {
		e.Scope = "business"
		e.Code = failedRes.Code
		e.Message = failedRes.Message
		return GetAccessTokenRes{}, e
	}

	e.Scope = "api"
	e.Message = "无法解析QQBot API返回内容: " + string(bodyBytes)
	return GetAccessTokenRes{}, e
}

// https://bot.q.qq.com/wiki/develop/api-v2/dev-prepare/api-call-guide.html
type APIFailedRes struct {
	ErrCode json.Number `json:"err_code"`
	Message string      `json:"message"`
	TraceID string      `json:"trace_id"`
}

func APIStatusCode(code int) string {
	switch code {
	case 200, 204:
		return "OK"
	case 201, 202:
		return "异步操作成功，请处理error body"
	case 401:
		return "未认证"
	case 404:
		return "没有该API"
	case 405:
		return "方法不被允许"
	case 429:
		return "请求过于频繁"
	case 500, 504:
		return "服务器错误"
	default:
		return "Unknown Status Code"
	}
}

// https://bot.q.qq.com/wiki/develop/api-v2/autogen/api/v2_groups_group_openid_messages.post.html
const API_SEND_GROUP_MESSAGE = QQBOT_API_BASE + "/v2/groups/{group_openid}/messages"

type SendGroupMessageReq struct {
	MsgType int    `json:"msg_type"` //一般都是0
	Content string `json:"content"`
}

type SendGroupMessageRes struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ExtInfo   struct {
		RefIdx string `json:"ref_idx"`
	} `json:"ext_info"`
}

func SendGroupMessage(req SendGroupMessageReq) (SendGroupMessageRes, error) {

	r, ctx, e := initRequest(http.MethodPost, API_SEND_GROUP_MESSAGE)
	_ = ctx

	groupOpenID, err := server.KVGet("OBSERVE_QQ_BOT_GROUP_OPENID")
	if err != nil {
		e.Scope = "client"
		e.Message = "获取GroupOpenID失败: " + err.Error()
		return SendGroupMessageRes{}, e
	}
	r.URL.Path = strings.Replace(r.URL.Path, "{group_openid}", groupOpenID, 1)

	request, err := json.Marshal(req)
	if err != nil {
		e.Scope = "client"
		e.Message = "JSON序列化请求失败: " + err.Error()
		return SendGroupMessageRes{}, e
	}
	r.Body = io.NopCloser(strings.NewReader(string(request)))
	res, err := c.Do(r)
	if err != nil {
		e.Scope = "client"
		e.Message = "请求发送失败: " + err.Error()
		return SendGroupMessageRes{}, e
	}
	defer res.Body.Close()

	var successRes SendGroupMessageRes
	var failedRes APIFailedRes

	bodyBytes, _ := io.ReadAll(res.Body)
	if err := json.Unmarshal(bodyBytes, &successRes); err == nil && successRes.ID != "" {
		return successRes, nil
	}

	if res.StatusCode != http.StatusOK {

		if err := json.Unmarshal(bodyBytes, &failedRes); err == nil && failedRes.ErrCode != "" {
			e.Scope = "business"
			e.Code = failedRes.ErrCode
			e.Message = failedRes.Message
			return SendGroupMessageRes{}, e
		}

		e.Scope = "api"
		e.Code = json.Number(fmt.Sprintf("%d", res.StatusCode))
		bodyBytes, _ := io.ReadAll(res.Body)
		e.Message = "QQBot API返回错误: " + e.Code.String() + ": " + APIStatusCode(res.StatusCode) + " " + fmt.Sprintf("%s", bodyBytes)
		return SendGroupMessageRes{}, e
	}

	e.Scope = "api"
	e.Message = "无法解析QQBot API返回内容: " + string(bodyBytes)
	return SendGroupMessageRes{}, e
}
