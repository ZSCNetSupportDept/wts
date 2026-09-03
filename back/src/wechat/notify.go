package wechat

import (
	"github.com/silenceper/wechat/v2/officialaccount"
	"github.com/silenceper/wechat/v2/officialaccount/message"
)

// SendNotify() 向指定用户发送一条微信通知。
//
// 该函数同时兼容微信现行的「订阅通知」（/cgi-bin/message/subscribe/bizsend）
// 与旧的「模板消息」（/cgi-bin/message/template/send）两套接口。由 wait 参数决定使用哪一套：
//
//	wait == true  : 发送订阅通知（用户需事先在 H5 页面授权，一次性额度）
//	wait == false : 发送旧模板消息（备用，微信政策变动时回退用，仅需调用处改参数）
//
// data 的 key 必须与所选模板的关键词字段名一致（如 thing1、phrase2 等）。
// url 为点击通知后跳转的页面绝对地址（域名须在公众号网页授权域名内）。
// 发送失败（含 43101 用户未订阅/额度耗尽/已取关）时返回 error，由调用方决定如何处理。
func SendNotify(wx *officialaccount.OfficialAccount, openid string, templateID string, data map[string]string, url string, wait bool) error {

	// 订阅通知结构
	subData := make(map[string]*message.SubscribeDataItem, len(data))
	for k, v := range data {
		subData[k] = &message.SubscribeDataItem{Value: v}
	}
	subMsg := &message.SubscribeMessage{
		ToUser:     openid,
		TemplateID: templateID,
		Page:       url,
		Data:       subData,
	}

	// 旧模板消息结构（备用）
	tplData := make(map[string]*message.TemplateDataItem, len(data))
	for k, v := range data {
		tplData[k] = &message.TemplateDataItem{Value: v}
	}
	tplMsg := &message.TemplateMessage{
		ToUser:     openid,
		TemplateID: templateID,
		URL:        url,
		Data:       tplData,
	}

	if wait {
		return wx.GetSubscribe().Send(subMsg)
	}
	_, err := wx.GetTemplate().Send(tplMsg)
	return err
}
