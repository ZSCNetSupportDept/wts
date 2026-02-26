package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/silenceper/wechat/v2/officialaccount"
	"zsxyww.com/wts/config"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
)

func Setup(cfg *config.Config, dbx *pgxpool.Pool, wx *officialaccount.OfficialAccount) *echo.Echo {
	app := echo.New()

	setDefaultContext(cfg, dbx, wx) // For custom context,read the comment on this function
	middlewareRegister(app, cfg)
	routeRegister(app, cfg)

	hutil.InitJWTKey(cfg.JWTKey)

	v := validator.New()
	hutil.RegisterValidator(v)
	app.Validator = &WtsValidator{validator: v}

	return app
}

type WtsValidator struct {
	validator *validator.Validate
}

func (w *WtsValidator) Validate(i interface{}) error {
	if err := w.validator.Struct(i); err != nil {
		return err
	}
	return nil
}
