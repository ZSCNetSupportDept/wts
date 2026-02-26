package handler

import (
	"log/slog"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/handler/logic"
	"zsxyww.com/wts/model/sqlc"
)

// POST: /api/v3/cancel_ticket
// receive: an URL parameter
// return 200 on success,400/403/500 on error
// type: JSON
func CancelTicket(i echo.Context) error {
	c := i.(*hutil.WtsCtx)

	var res hutil.CancelTicketResponse

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	var u *hutil.WtsJWT

	var isCancelingSelf bool

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

	tid := i.QueryParam("tid")
	if tid == "" {
		res.Success = false
		res.ErrType = hutil.ErrReq
		res.Msg = "missing required URL parameter: tid"
		return i.JSON(400, res)
	}

	//校验权限
	ta, err := strconv.Atoi(tid)
	if err != nil {
		res.Success = false
		res.ErrType = hutil.ErrReq
		res.Msg = "invalid ticket ID: " + err.Error()
		return i.JSON(400, res)
	}
	tidInt := int32(ta)

	if u.Access == sqlc.WtsAccessDev {
		slog.Info("cancel_ticket:检测到最高权限，跳过工单归属检查", "id", id, "tid", tidInt, "op", u.OpenID)
		goto do
	}

	isCancelingSelf, err = logic.IsOwningTicket(c, u.Sid, tidInt)
	if err != nil {
		res.Success = false
		res.ErrType = hutil.ErrReq
		res.Msg = "cannot fetch ticket info: " + err.Error()
		return i.JSON(400, res)
	}
	if !isCancelingSelf {
		res.Success = false
		res.ErrType = hutil.ErrAuth
		res.Msg = "you can only cancel tickets of your own"
		return i.JSON(403, res)
	}

do:
	r := logic.AppendTraceParam{
		Tid:            tidInt,
		NewStatus:      sqlc.WtsStatusCanceled,
		NewPriority:    "",
		NewAppointment: time.Time{},
		Remark:         "已自行取消报修",
	}

	res.Err = logic.AppendTrace(c, u.OpenID, r)
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
		res.Msg = "ticket cancel success~"
		slog.Debug("请求成功返回", "id", id)
	}

	slog.Debug("原始返回内容:", "id", id, "content", res)
	return i.JSON(res.Code, res)
}
