package server

import (
	"github.com/labstack/echo/v4"
	"zsxyww.com/wts/config"
	"zsxyww.com/wts/handler"
)

func routeRegister(app *echo.Echo, cfg *config.Config) { // Routes

	// Static Files,FrontEnd
	app.Static("/", cfg.FrontEndDir)

	// Groups
	api := app.Group("/api")
	v3 := api.Group("/v3") // The system is the version 3 of our Wechat Ticket System
	v3p := api.Group("/v3p")
	v3rest := v3.Group("/rest")

	if !cfg.Debug.SkipJWTAuth {
		v3.Use(JWTAuthMiddleware(cfg.JWTKey))
		v3rest.Use(JWTAuthMiddleware(cfg.JWTKey))
	}

	{ //Debug and Miscellaneous Routes
		app.GET("/br-debug/testdb", handler.TestDB)
		app.GET("/br-debug/panic", handler.Panic)

		app.GET("api", handler.Hello)
		api.GET("/", handler.Hello)
		api.GET("/v3", handler.Hello)
		api.GET("/v3/", handler.Hello)
		api.GET("/v3/rest", handler.Hello)
		api.GET("/v3/rest/", handler.Hello)
	}

	{ // Business Windows
		v3.POST("/register", handler.Register)
		v3.POST("/change_profile", handler.ChangeProfile)
		v3.GET("/view_profile", handler.ViewProfile)
		v3.POST("/filter_users", handler.FilterUsers)
		v3.POST("/new_ticket", handler.NewTicket)
		v3.GET("/get_ticket", handler.GetTicket)
		v3.POST("/cancel_ticket", handler.CancelTicket)
		v3.POST("/new_repair_trace", handler.NewRepairTrace)
		v3.POST("/filter_tickets", handler.FilterTickets)
		v3.GET("/get_traces", handler.GetTraces)
		v3.GET("/ticket_overview", handler.TicketOverview)

	}

	{ //RESTful Resources API
		v3rest.GET("/test", handler.Hello)
	}

	{ //WeChat Server Communication API
		v3p.GET("/wx", handler.WXEntry)
		v3p.POST("/wx", handler.WXEntry)
		v3p.GET("/wx/auth", handler.WXAuth)
		v3p.GET("/wx/authsuccess", handler.WXAuthSuccess)
	}

}
