package handler

import (
	"log/slog"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/handler/logic"
	"zsxyww.com/wts/model/sqlc"
)

// GET: /api/v3/view_profile
// receive: an URL parameter
// return 200 on success,400/403/500 on error
// type: JSON
func ViewProfile(i echo.Context) error {
	c := i.(*hutil.WtsCtx)

	var res hutil.ViewUserProfileResponse

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	var u *hutil.WtsJWT

	slog.Info("收到HTTP请求", "id", id, "URI", i.Request().URL.String(), "from", i.RealIP(), "method", i.Request().Method, "user_agent", i.Request().UserAgent())
	slog.Debug("具体的请求信息", "id", id, "headers", i.Request().Header, "query_params", i.Request().URL.Query(), "body", i.Get("body"))

	//校验权限
	if !c.Cfg.Debug.SkipJWTAuth {
		u = i.Get("jwt").(*jwt.Token).Claims.(*hutil.WtsJWT)
		if !hutil.IsUser(u.Access) {
			res.Success = false
			res.ErrType = hutil.ErrAuth
			res.Msg = "only active users can access this API"
			return i.JSON(403, res)
		}
		slog.Debug("鉴权已通过", "id", id, "Content", u)
	} else {
		slog.Info("已跳过JWT验证", "id", id)
		u = &hutil.WtsJWT{OpenID: "system", Access: sqlc.WtsAccessDev}
	}

	who := i.QueryParam("who")
	if who == "" {
		who = u.OpenID
	}

	if u.OpenID != who {
		if !hutil.IsAdmin(u.Access) {
			res.Success = false
			res.ErrType = hutil.ErrAuth
			res.Msg = "only admins can view other users' profiles"
			return i.JSON(403, res)
		}
	}

	res = logic.ViewProfile(c, u.OpenID, who)
	if res.Err != nil {
		res.Success = false
		res.ErrType = hutil.ErrLogic
		if hutil.IsKnownErr(res.Err) {
			res.Code = 400
			res.Msg = res.Err.Error()
			if c.Cfg.Debug.APIVerbose {
				res.Debug = res.Err.(*hutil.WtsErr).Unwrap().Error()
			}
			slog.Info("请求出现已捕获错误", "id", id, "error", res.Err)
		} else {
			res.Code = 500
			res.Msg = "system met a uncaught error,please view logs."
			res.ErrType = hutil.ErrInternal
			if c.Cfg.Debug.APIVerbose {
				res.Debug = res.Err.Error()
			}
			slog.Warn("请求出现未捕获错误！", "id", id, "error", res.Err)
		}
	} else {
		res.Success = true
		res.Code = 200
		res.Msg = "user profile"
		slog.Debug("请求成功返回", "id", id)
	}

	return i.JSON(res.Code, res)

}
