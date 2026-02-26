// 和微信有关的handler
package handler

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/handler/logic"
	"zsxyww.com/wts/model/sqlc"
)

// GET: /api/v3p/wx
// POST: /api/v3p/wx
// reveive: WeChat Specificed Request
// return 200 on success,500 on error
// type: WeCaht Specificed JSON/XML Response
// this API is used to communicate with WeChat Server, Frontend developers are unneeded to care about it.
func WXEntry(i echo.Context) error {
	c := i.(*hutil.WtsCtx)

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	slog.Info("收到HTTP请求", "id", id, "URI", i.Request().URL.Path, "from", i.RealIP(), "method", i.Request().Method, "user_agent", i.Request().UserAgent())
	slog.Debug("具体的请求信息", "id", id, "headers", i.Request().Header, "query_params", i.Request().URL.Query(), "body", i.Get("body"))

	wc := c.WX.GetServer(i.Request(), i.Response().Writer)
	//负责回复用户从公众号聊天栏发来的消息与推送的event
	wc.SetMessageHandler(logic.WXMsgHandler(c))
	err := wc.Serve()
	if err != nil {
		i.Logger().Error("wechat server error:", err)
		i.String(500, "in: "+time.Now().String()+" wechat handler error,please view logs.")
		return err
	}
	wc.Send()
	return nil
}

// 执行微信的OAuth2.0授权流程，跳转到微信授权页面
// GET: /api/v3p/wx/auth
// receive: none
// return: 500 on error
// type: string on error,no return on success, passing a cookie for OAuth2.0 verification
// special: redirect to WeChat OAuth2.0 authorization page
func WXAuth(i echo.Context) error {
	c := i.(*hutil.WtsCtx)

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	slog.Info("收到HTTP请求", "id", id, "URI", i.Request().URL.Path, "from", i.RealIP(), "method", i.Request().Method, "user_agent", i.Request().UserAgent())
	slog.Debug("具体的请求信息", "id", id, "headers", i.Request().Header, "query_params", i.Request().URL.Query(), "body", i.Get("body"))

	// 构造请求参数
	uri := "https://" + c.Cfg.WX.CallBackURL + "/api/v3p/wx/authsuccess"
	s := "snsapi_userinfo"
	state := genAuthState()

	// 将随机生成的state存到前端，用来在回调时同步校验
	cookie := new(http.Cookie)
	cookie.Name = "oauth_state"
	cookie.Path = "/"
	cookie.Value = state
	cookie.Expires = time.Now().Add(5 * time.Minute)
	cookie.HttpOnly = true
	cookie.Secure = true
	i.SetCookie(cookie)

	// 重定向到微信授权页面
	to, err := c.WX.GetOauth().GetRedirectURL(uri, s, state)
	if err != nil {
		return c.String(500, "生成WeChat OAuth2.0授权链接失败:"+err.Error())
	}
	return c.Redirect(http.StatusFound, to)
}

// GET: /api/v3p/wx/authsuccess
// receive: WeChat specificed OAuth2.0 callback parameters
// return: 400/500 on error,200 on success
// type: string on error, Redirect on success(Cookie containing JWT is set)
// this is automaticly accessed after OAuth2.0 authorization success.
func WXAuthSuccess(i echo.Context) error {
	c := i.(*hutil.WtsCtx)

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	slog.Info("收到HTTP请求", "id", id, "URI", i.Request().URL.Path, "from", i.RealIP(), "method", i.Request().Method, "user_agent", i.Request().UserAgent())
	slog.Debug("具体的请求信息", "id", id, "headers", i.Request().Header, "query_params", i.Request().URL.Query(), "body", i.Get("body"))

	// 校验微信返回
	s := i.QueryParam("state")
	if s == "" {
		return c.String(http.StatusBadRequest, "未获取到授权 state")
	}

	cookie, err := c.Cookie("oauth_state")
	if err != nil {
		return c.String(http.StatusBadRequest, "无法获取 state cookie，请求可能已过期或非法")
	}

	if s != cookie.Value {
		return c.String(http.StatusForbidden, "state 参数校验失败")
	}

	cookie.Expires = time.Now().Add(-1 * time.Hour)
	i.SetCookie(cookie)

	code := i.QueryParam("code")
	if code == "" {
		return c.String(http.StatusBadRequest, "未获取到授权 code")
	}

	res, err := c.WX.GetOauth().GetUserAccessToken(code)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("换取 access_token 失败: %v", err))
	}

	openID := res.OpenID

	userInfo, err := c.WX.GetOauth().GetUserInfo(res.AccessToken, openID, "zh_CN")
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("获取用户信息失败: %v", err))
	}

	// 查询数据库，看看有没有OpenID对应的用户
	ctx := i.Request().Context()
	u := sqlc.WtsVUser{}
	var reg bool
	if err = c.DB.DoQuery(ctx, openID, func(q *sqlc.Queries) error {
		u, err = q.GetUserByWX(ctx, openID)
		if err != nil {
			switch true {
			case errors.Is(err, pgx.ErrNoRows):
				reg = false
				return nil
			default:
				return err
			}
		}
		reg = true
		return nil
	}); err != nil {
		return c.String(500, "Database Query Error,Transaction Rollbacked:"+err.Error())
	}

	//生成对应的JWT
	var t string
	if !reg {
		t, err = hutil.NewWtsJWT(openID, "", sqlc.WtsAccessUnregistered, userInfo.Nickname, userInfo.HeadImgURL, "用户", 30)
		if err != nil {
			return c.String(500, "生成JWT(临时)失败:"+err.Error())
		}
	} else {
		var access sqlc.WtsAccess
		if u.Op && u.Access.Valid {
			access = u.Access.WtsAccess
		} else {
			access = sqlc.WtsAccessUser
		}
		t, err = hutil.NewWtsJWT(u.Wx, u.Sid.String, access, userInfo.Nickname, userInfo.HeadImgURL, emptyText1(u.Name), 95)
		if err != nil {
			return c.String(500, "JWT生成失败："+err.Error())
		}
	}

	// 将JWT写入Cookie，后端不从cookie读取JWT，前端应该立即将该cookie存储到localStorage并通过请求头传递JWT
	jwt := new(http.Cookie)
	jwt.Name = "jwt"
	jwt.Value = t
	jwt.Expires = time.Now().Add(5 * time.Minute)
	jwt.HttpOnly = false
	jwt.Secure = true
	jwt.Path = "/"
	i.SetCookie(jwt)

	return c.Redirect(http.StatusFound, c.Cfg.FrontEnd.OnAuthSuccess)
}

func genAuthState() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := [16]byte{}
	for i := 0; i < 16; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result[:])
}

func emptyText1(t pgtype.Text) string {
	if !t.Valid {
		return "你是谁？"
	}
	return t.String
}
