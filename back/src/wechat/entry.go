package wechat

import (
	wx "github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	"github.com/silenceper/wechat/v2/officialaccount"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"zsxyww.com/wts/config"
)

var Memory *cache.Memory

func Setup(cfg *config.Config) *officialaccount.OfficialAccount {

	wc := wx.NewWechat()

	//TODO:从.dat文件加载access token or 使用Redis等其他缓存方式

	Memory = cache.NewMemory()

	wxcfg := &offConfig.Config{
		AppID:          cfg.WX.AppID,
		AppSecret:      cfg.WX.AppSecret,
		Token:          cfg.WX.Token,
		EncodingAESKey: cfg.WX.EncodingAESKey,
		Cache:          Memory,
	}

	officialAccount := wc.GetOfficialAccount(wxcfg)

	return officialAccount
}
