package handler

import (
	"log/slog"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/handler/logic"
	"zsxyww.com/wts/model/sqlc"
)

// POST: /api/v3/new_repair_trace
// receive: JSON
// return 200 on success,400/403/500 on error
// type: JSON
func NewRepairTrace(i echo.Context) error {
	c := i.(*hutil.WtsCtx)

	var res hutil.CancelTicketResponse

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	var u *hutil.WtsJWT

	slog.Info("收到HTTP请求", "id", id, "URI", i.Request().URL.String(), "from", i.RealIP(), "method", i.Request().Method, "user_agent", i.Request().UserAgent())
	slog.Debug("具体的请求信息", "id", id, "headers", i.Request().Header, "query_params", i.Request().URL.Query(), "body", i.Get("body"))

	//校验权限
	if !c.Cfg.Debug.SkipJWTAuth {
		u = i.Get("jwt").(*jwt.Token).Claims.(*hutil.WtsJWT)
		if !hutil.IsOperator(u.Access) {
			res.Success = false
			res.ErrType = hutil.ErrAuth
			res.Msg = "only Network Support staff can access this API"
			return i.JSON(403, res)
		}
		slog.Debug("鉴权已通过", "id", id, "Content", u)
	} else {
		slog.Info("已跳过JWT验证", "id", id)
		u = &hutil.WtsJWT{OpenID: "system", Access: sqlc.WtsAccessDev}
	}

	r := hutil.NewRepairTraceRequest{}
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

	//校验权限
	if !(r.NewAppointment.IsZero()) && r.NewStatus != string(sqlc.WtsStatusScheduled) {
		res.Success = false
		res.ErrType = hutil.ErrReq
		res.Msg = "only appointed status can set appointment time"
		return i.JSON(400, res)
	}

	if r.NewPriority != "" {
		if !hutil.IsAdmin(u.Access) {
			res.Success = false
			res.ErrType = hutil.ErrAuth
			res.Msg = "only admin can change ticket priority"
			return i.JSON(403, res)
		}
	}

	ra := logic.AppendTraceParam{
		Tid:            r.Tid,
		NewStatus:      sqlc.WtsStatus(r.NewStatus),
		NewPriority:    sqlc.WtsPriority(r.NewPriority),
		NewAppointment: r.NewAppointment,
		Remark:         r.Remark,
	}

	res.Err = logic.AppendTrace(c, u.OpenID, ra)
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
		res.Msg = "add trace to ticket success~"
		slog.Debug("请求成功返回", "id", id)
	}

	slog.Debug("原始返回内容:", "id", id, "content", res)
	return i.JSON(res.Code, res)

}
