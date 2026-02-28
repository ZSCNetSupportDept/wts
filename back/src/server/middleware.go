package server

import (
	"strings"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"zsxyww.com/wts/config"
	. "zsxyww.com/wts/handler/handlerUtilities"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
)

func middlewareRegister(app *echo.Echo, cfg *config.Config) {
	app.Use(customContext)
	app.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Skipper: middleware.DefaultSkipper,
	}))
	if cfg.JSONLogOutput {
		app.Use(middleware.LoggerWithConfig(json))
	} else {
		app.Use(middleware.LoggerWithConfig(human2))
		_ = human
	}
	app.Use(middleware.Recover())
	app.Use(middleware.Secure())
	app.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20.0)))

	app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Path(), "/api")
		},
	}))
}

func customContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := &WtsCtx{
			Context: c,
			Cfg:     Cfg,
			DBx:     DBx,
			DB:      DB,
			WX:      WX,
		}
		return next(cc)
	}
}

func JWTAuthMiddleware(key string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(hutil.WtsJWT)
		},
		SigningKey: []byte(key),
		ContextKey: "jwt",
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(403, map[string]string{
				"msg":        "JWT没有找到，或内容无效。",
				"err":        err.Error(),
				"success":    "false",
				"error_type": string(rune(hutil.ErrAuth)),
			})
		},
	})
}

var human = middleware.LoggerConfig{
	Skipper: middleware.DefaultSkipper,
	Format: `${time_custom} [Info] [Echo] HTTP Request Received:` +
		`"${method} ${uri}"from ${remote_ip};` +
		`Respond With:${status} ${error} in ${latency_human};` +
		`UA:${user_agent},bytes_in:${bytes_in},bytes_out:${bytes_out},ID:${id}` + "\n",
	CustomTimeFormat: "2006-01-02 15:04:05.00000",
}

var json = middleware.LoggerConfig{
	Skipper: middleware.DefaultSkipper,
	Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
		`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
		`"status":${status},"error":"${error}","latency":${latency},"latency_human":"${latency_human}"` +
		`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
	CustomTimeFormat: "2006-01-02 15:04:05.00000",
}

var human2 = middleware.LoggerConfig{
	Skipper: middleware.DefaultSkipper,
	Format: `time=${time_custom} level=INFO msg=HTTP请求已完成 ` +
		`uri="${method} ${uri}" from=${remote_ip} user_agent="${user_agent}" ` +
		`id=${id} respond=${status} latency=${latency_human} error(if do exist)=${error} ` +
		`bytes_in=${bytes_in} bytes_out=${bytes_out} ` + "\n",
	CustomTimeFormat: "2006-01-02T15:04:05.000+00:00",
}
