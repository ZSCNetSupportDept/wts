package hutil

type CommonErr int

const (
	ErrInternal CommonErr = iota + 1 //服务器内部错误
	ErrReq                           //无效请求
	ErrAuth                          //未授权访问
	ErrDB                            //数据库错误
	ErrLogic                         // 业务逻辑错误
)

var commonErrMsg = map[CommonErr]string{
	CommonErr(0): "无错误",
	ErrInternal:  "服务器内部错误",
	ErrReq:       "你的请求无效",
	ErrAuth:      "鉴权失败了",
	ErrDB:        "数据库出现错误！",
	ErrLogic:     "业务逻辑出现错误！",
}

func (e CommonErr) Error() string {
	if msg, exists := commonErrMsg[e]; exists {
		return msg
	}

	return "unknown error" + string(rune(e))
}
