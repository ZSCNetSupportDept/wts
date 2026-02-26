package server

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/silenceper/wechat/v2/officialaccount"
	"zsxyww.com/wts/config"
	"zsxyww.com/wts/model"
)

var Cfg *config.Config
var DBx *pgxpool.Pool
var DB *model.Store
var WX *officialaccount.OfficialAccount

// 为handler传递自定义上下文分成如下几步：
//
// 1.修改WtsCtx结构体，添加需要传递的变量
//
// 2.将变量作为参数传递给server.Setup函数并进一步传递给本函数，在这段话的上面声明全局变量，在下面的函数里为他们赋值
//
// 3.转到server.customContext函数，修改cc的赋值，使其包含上面的变量
//
// 然后，你就可以在handler中通过类型断言获取这些变量了 ，例如：c:= i.(*WtsCtx)
// 写测试的时候，只需要手动构造WtsCtx，即可
var setDefaultContext = func(cfg *config.Config, dbx *pgxpool.Pool, wx *officialaccount.OfficialAccount) error {
	Cfg = cfg
	DBx = dbx
	DB = model.NewStore(dbx)
	WX = wx
	return nil
}
