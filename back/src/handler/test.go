package handler

import (
	"context"

	"github.com/labstack/echo/v4"
	. "zsxyww.com/wts/handler/handlerUtilities"
)

// GET: /br-debug/testdb
// reveive: none
// return 200 on success,500 on error
// type: string
func TestDB(i echo.Context) error {
	c := i.(*WtsCtx)

	if err := c.DBx.Ping(context.Background()); err != nil {
		return i.String(500, "database test error:"+err.Error())
	}
	return i.String(200, "Database connection is healthy")

}

// GET: /api/
// GET: /api/v3/
// GET: /api/v3/rest/
// GET: /api/v3/rest/test
// etc.
// reveive: none
// return 200 on success
// type: string
func Hello(i echo.Context) error {
	c := i.(*WtsCtx)
	brand := c.Cfg.Brand
	return i.String(200, "Welcome to "+brand+",For more information, please visit http://www.zsxyww.com/wtsdocs")
}

func Panic(i echo.Context) error {
	panic("this is a test panic")
}
