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

	//然后，启动服务器
	app := server.Setup(cfg, dbx, wx)

	//启动守护进程（因为有的服务需要用到上下文，所以现在修改在server.Setup被执行的后面启动）
	daemon.Setup()

	err := app.Start("127.0.0.1:" + strconv.Itoa(cfg.ListenPort))

	println("Server exited." + err.Error())
}
