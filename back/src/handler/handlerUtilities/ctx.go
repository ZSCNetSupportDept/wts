package hutil

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/silenceper/wechat/v2/officialaccount"
	"zsxyww.com/wts/config"
	"zsxyww.com/wts/model"
)

type WtsCtx struct {
	echo.Context
	Cfg *config.Config
	DBx *pgxpool.Pool
	DB  *model.Store
	WX  *officialaccount.OfficialAccount
}
