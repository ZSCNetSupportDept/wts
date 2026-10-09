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

// POST: /api/v3/change_appointment
// receive: JSON, view docs
// return 200 on success,400/403/500 on error
// type: JSON
// ChangeAppointment 让用户自己可以修改工单的预约时间，只能修改自己提交的工单。
func ChangeAppointment(i echo.Context) error {

	c := i.(*hutil.WtsCtx)

	var res hutil.ChangeAppointmentResponse

	id := i.Response().Header().Get(echo.HeaderXRequestID)

	var u *hutil.WtsJWT

	var isChangingSelf bool

	var t sqlc.WtsVTicket

	var err error

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

	//校验并绑定请求体的数据
	r := hutil.ChangeAppointmentRequest{}
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

	if !r.NewAppointment.IsZero() && r.NewAppointment.Before(time.Now()) { //TODO:预约可以在今天！！！重要！！！
		res.Success = false
		res.ErrType = hutil.ErrReq
		res.Msg = "appointment time cannot be in the past"
		return i.JSON(400, res)
	}

	if u.Access == sqlc.WtsAccessDev {
		slog.Info("change_appointment:检测到最高权限，跳过工单归属检查", "id", id, "tid", r.Tid, "op", u.OpenID)
		goto do
	}

	t, err = logic.GetTicketByTID(c, r.Tid)
	if err != nil {
		res.Success = false
		res.ErrType = hutil.ErrReq
		res.Msg = "cannot fetch ticket info: " + err.Error()
		return i.JSON(400, res)
	}

	isChangingSelf = (t.Issuer == u.Sid)

	if !isChangingSelf {
		res.Success = false
		res.ErrType = hutil.ErrAuth
		res.Msg = "you can only change appointments of your own tickets"
		return i.JSON(403, res)
	}

	if !(t.Status == sqlc.WtsStatusScheduled ||
		t.Status == sqlc.WtsStatusDelay ||
		t.Status == sqlc.WtsStatusFresh) {
		res.Success = false
		res.ErrType = hutil.ErrReq
		res.Msg = "cannot change appointment for tickets that are not scheduled, delayed, or fresh"
		return i.JSON(400, res)
	}

do:
	r1 := logic.AppendTraceParam{
		Tid:            r.Tid,
		NewStatus:      sqlc.WtsStatusScheduled,
		NewPriority:    sqlc.WtsPriority(""),
		NewAppointment: r.NewAppointment,
		Remark:         "已自行变更预约",
	}

	res.Err = logic.AppendTrace(c, u.OpenID, r1)
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
		res.Msg = "change appointment success~"
		res.Tid = r.Tid
		res.NewAppointment = r.NewAppointment
		slog.Debug("请求成功返回", "id", id)
	}

	slog.Debug("原始返回内容:", "id", id, "content", res)
	return i.JSON(res.Code, res)

}
