package handler

import (
	"log/slog"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
)

func RevokeMyToken(i echo.Context) error {
	c := i.(*hutil.WtsCtx)

	id := i.Response().Header().Get(echo.HeaderXRequestID)
	slog.Info("收到HTTP请求", "id", id, "URI", i.Request().URL.Path, "from", i.RealIP())

	user := i.Get("jwt")
	if user == nil {
		return i.JSON(403, map[string]string{
			"msg":     "无法获取登录凭证。",
			"success": "false",
		})
	}
	token, ok := user.(*jwt.Token)
	if !ok {
		return i.JSON(500, map[string]string{
			"msg":     "服务器内部错误。",
			"success": "false",
		})
	}
	claims, ok := token.Claims.(*hutil.WtsJWT)
	if !ok {
		return i.JSON(500, map[string]string{
			"msg":     "服务器内部错误。",
			"success": "false",
		})
	}

	if claims.ID == "" {
		return i.JSON(400, map[string]string{
			"msg":     "当前token没有jti，无法吊销。",
			"success": "false",
		})
	}

	hutil.TokenBl.Add(claims.ID, claims.ExpiresAt.Time)
	slog.Info("JWT已被服务端吊销", "id", id, "jti", claims.ID, "openid", claims.OpenID)

	return i.JSON(200, map[string]string{
		"msg":     "登录凭证已吊销，后续请求将被拒绝。请重新登录。",
		"success": "true",
	})
}
