package main

import (
	"strconv"

	"zsxyww.com/wts/config"
	"zsxyww.com/wts/daemon"
	"zsxyww.com/wts/db"
	"zsxyww.com/wts/logger"
	"zsxyww.com/wts/server"
	"zsxyww.com/wts/wechat"
)

func main() {

	//首先，加载所需的配置文件
	cfg := config.Load()

	//再初始化日志模块(slog)
	logger.Setup(cfg)

	//其次，连接数据库
	dbx := db.Connect(cfg)

	//设置微信SDK
	wx := wechat.Setup(cfg)

	//启动守护进程
	daemon.Setup()

	//然后，启动服务器
	app := server.Setup(cfg, dbx, wx)
	err := app.Start("127.0.0.1:" + strconv.Itoa(cfg.ListenPort))

	println("Server exited." + err.Error())
}
