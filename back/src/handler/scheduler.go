package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/scheduler"
)

// JWT verifies OpenID; scheduler.Service resolves the current role from the database.
func Scheduler(action string) echo.HandlerFunc {
	return func(i echo.Context) error {
		c := i.(*hutil.WtsCtx)
		token, ok := i.Get("jwt").(*jwt.Token)
		if !ok || !token.Valid {
			return i.JSON(403, map[string]any{"success": false, "error_type": 3, "msg": "排班接口要求有效JWT，不支持跳过鉴权"})
		}
		claims, ok := token.Claims.(*hutil.WtsJWT)
		if !ok || claims.OpenID == "" {
			return i.JSON(403, map[string]any{"success": false, "error_type": 3, "msg": "无效登录身份"})
		}
		service := scheduler.Service{Pool: c.DBx}
		ctx := i.Request().Context()
		var result any
		var err error
		switch action {
		case "overview":
			result, err = service.Overview(ctx, claims.OpenID, i.QueryParam("month"))
		case "staff":
			result, err = service.Staff(ctx, claims.OpenID, i.QueryParam("date"))
		case "audit":
			result, err = service.Audit(ctx, claims.OpenID, i.QueryParam("date"))
		case "swaps":
			result, err = service.Swaps(ctx, claims.OpenID, i.QueryParam("month"))
		default:
			i.Request().Body = http.MaxBytesReader(i.Response(), i.Request().Body, 65536)
			switch action {
			case "availability":
				var r scheduler.AvailabilityRequest
				if err = i.Bind(&r); err == nil {
					result, err = service.Availability(ctx, claims.OpenID, r)
				} else {
					err = &scheduler.Error{Code: 400, Message: "请求内容格式错误"}
				}
			case "swap_request", "swap_review":
				var r scheduler.SwapRequest
				if err = i.Bind(&r); err == nil {
					if action == "swap_request" {
						result, err = service.CreateSwap(ctx, claims.OpenID, r)
					} else {
						result, err = service.ReviewSwap(ctx, claims.OpenID, r)
					}
				} else {
					err = &scheduler.Error{Code: 400, Message: "请求内容格式错误"}
				}
			default:
				var r scheduler.Request
				if err = i.Bind(&r); err == nil {
					result, err = service.Change(ctx, claims.OpenID, action, r)
				} else {
					err = &scheduler.Error{Code: 400, Message: "请求内容格式错误"}
				}
			}
		}
		if err != nil {
			var e *scheduler.Error
			if errors.As(err, &e) {
				kind := 2
				if e.Code == 403 {
					kind = 3
				}
				return i.JSON(e.Code, map[string]any{"success": false, "error_type": kind, "msg": e.Message})
			}
			slog.Error("排班请求失败", "request_id", i.Response().Header().Get(echo.HeaderXRequestID), "action", action, "error", err)
			return i.JSON(500, map[string]any{"success": false, "error_type": 1, "msg": "排班服务暂不可用，请联系管理员检查数据库迁移"})
		}
		return i.JSON(200, map[string]any{"success": true, "data": result})
	}
}
