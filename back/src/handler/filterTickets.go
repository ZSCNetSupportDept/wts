package handler

import (
	"log/slog"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/handler/logic"
	"zsxyww.com/wts/model/sqlc"
)

// POST: /api/v3/filter_tickets
// receive: JSON,view docs
// return 200 on success,400/403/500 on error
// type: JSON
func FilterTickets(i echo.Context) error {

	c := i.(*hutil.WtsCtx)

	var res hutil.FilterTicketsResponse

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	var u *hutil.WtsJWT

	slog.Info("收到HTTP请求", "id", id, "URI", i.Request().URL.Path, "from", i.RealIP(), "method", i.Request().Method, "user_agent", i.Request().UserAgent())
	slog.Debug("具体的请求信息", "id", id, "headers", i.Request().Header, "query_params", i.Request().URL.Query(), "body", i.Get("body"))

	//校验权限
	if !c.Cfg.Debug.SkipJWTAuth {
		u = i.Get("jwt").(*jwt.Token).Claims.(*hutil.WtsJWT)
		if !hutil.IsOperator(u.Access) {
			res.Success = false
			res.Msg = "only staff can access this API"
			res.ErrType = hutil.ErrAuth
			return i.JSON(403, res)
		}
		slog.Debug("鉴权已通过", "id", id, "Content", u)
	} else {
		slog.Info("已跳过JWT验证", "id", id)
		u = &hutil.WtsJWT{OpenID: "system", Access: sqlc.WtsAccessDev}
	}

	//校验并绑定请求体的数据
	r := hutil.FilterTicketsRequest{}
	if err := i.Bind(&r); err != nil {
		slog.Info("请求体绑定失败", "id", id, "error", err)
		res.Success = false
		res.Msg = "cannot bind your request body: " + err.Error()
		res.ErrType = hutil.ErrReq
		return i.JSON(400, res)
	}
	if err := i.Validate(&r); err != nil {
		slog.Info("请求体验证失败", "id", id, "error", err)
		res.Success = false
		res.Msg = "invalid request body: " + err.Error()
		res.ErrType = hutil.ErrReq
		return i.JSON(400, res)
	}

	//处理Scope参数的合法性
	switch r.Scope {
	case "all":
		if !hutil.IsAdmin(u.Access) {
			res.Success = false
			res.Msg = "only admin can filter all tickets"
			res.ErrType = hutil.ErrAuth
			return i.JSON(403, res)
		}
	case "active":
		break
	default:
		res.Success = false
		res.Msg = "invalid scope value"
		res.ErrType = hutil.ErrReq
		return i.JSON(400, res)
	}
	//处理时间参数的合法性
	if r.OlderThan.IsZero() {
		r.OlderThan = time.Now()
	}
	if r.NewerThan.After(r.OlderThan) {
		res.Success = false
		res.Msg = "newerThan cannot be after olderThan"
		res.ErrType = hutil.ErrReq
		return i.JSON(400, res)
	}

	//调用逻辑层处理请求
	res = logic.FilterTickets(c, u.OpenID, r)
	// 处理返回结果
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
			slog.Error("请求出现未捕获错误", "id", id, "error", res.Err)
		}
		return i.JSON(res.Code, res)
	}

	res.Success = true
	res.Code = 200
	res.Msg = "query success"
	slog.Debug("请求成功返回", "id", id)

	slog.Debug("原始返回内容:", "id", id, "content", res)
	return i.JSON(res.Code, res)
}
