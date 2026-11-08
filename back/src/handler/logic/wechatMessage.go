package logic

import (
	"log/slog"

	"github.com/silenceper/wechat/v2/officialaccount/message"
	. "zsxyww.com/wts/handler/handlerUtilities"
)

// WXMsgHandler 处理微信消息的逻辑
// 微信的所有消息都通过一个入口送过来，对应这里的函数来集中处理
// 用户发过来的消息可以分成如下类别：
//  1. 事件消息（Event）
//     包括关注、取消关注、点击菜单等事件
//  2. 普通消息
//     包括文本消息、图片消息、语音消息等，目前只处理文本消息
//  3. 命令
//     以“/”开头的文本消息，作为系统命令来处理
func WXMsgHandler(c *WtsCtx) func(msg *message.MixMessage) *message.Reply {
	return func(msg *message.MixMessage) *message.Reply {

		i := Ctx(*c)

		if c.Cfg.WX.AIReply {
			reply, ok := i.handleAIReply(msg)
			if ok {
				return reply
			}
		}

		//处理事件
		if msg.MsgType == message.MsgTypeEvent {
			reply := message.NewText(i.handleWXEvent(msg))
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: reply}
		}

		// 处理用户发过来的信息和命令
		if msg.MsgType == message.MsgTypeText {
			reply := message.NewText(i.handleWXTxtMsg(msg))
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: reply}
		}

		if msg.MsgType != message.MsgTypeText {
			reply := message.NewText("Sorry ，系统目前只能处理文字消息~")
			return &message.Reply{MsgType: message.MsgTypeText, MsgData: reply}
		}

		return nil
	}
}

func (i *Ctx) handleWXEvent(m *message.MixMessage) string {
	//可以假定进入这里的信息全部是event类型

	// 用户关注时发送欢迎文本
	if m.Event == message.EventSubscribe {
		return "同学你好，欢迎使用网维报修系统~\n\n建议先看看使用攻略呢：https://wwbx.davisye.cn/help\n"
	}
	if m.Event == message.EventView {
		// 不知道为什么，view事件也会被送到这里，如果不处理的话会在log里面出现，有点烦，对用户体验倒是没什么影响
		return ""
	}

	// 订阅通知授权事件：记录微信通知的授权结果，只是用于观测用户的授权概率
	if m.Event == message.EventSubscribeMsgPopupEvent {
		for _, item := range m.GetSubscribeMsgPopupEvents() {
			slog.Info("订阅通知授权事件结果",
				"openid", m.GetOpenID(),
				"template_id", item.TemplateID,
				"status", item.SubscribeStatusString, // accept / reject / ban
				"scene", item.PopupScene,
				"event", m.EventKey,
			)
		}
		return ""
	}

	if m.Event == message.EventType("subscribe_msg_sent_event") {
		// 订阅通知成功发送时会送达该事件，不需要回复，否则会在通知后发送额外报错信息
		return ""
	}

	return "错误：事件" + string(m.Event) + "::" + m.EventKey + "没有被定义处理逻辑"

}

func (i *Ctx) handleWXTxtMsg(m *message.MixMessage) string {
	//检查是否以"/"开头
	if len(m.Content) > 0 && m.Content[0] == '/' {
		return i.processCommand(m)
	}

	//TODO: 处理普通文本消息
	return i.handleNormalMsg(m)
}

func (i *Ctx) handleNormalMsg(m *message.MixMessage) string {
	_ = i.superEasyNLPProgram(m)
	return "您好，系统暂时不支持自动回复哦~点击<a href=\"https://wwbx.davisye.cn\">这里</a>进入系统界面进行报修吧~"
	//return i.superEasyNLPProgram(m)
}

func (i *Ctx) superEasyNLPProgram(m *message.MixMessage) string {
	verb := string([]rune(m.Content)[0])
	return verb + verb + verb + "，一天到晚就™知道" + m.Content + "，是不是军姿没站够？"
}

func isTextACommand(text string) bool {
	return len(text) > 0 && text[0] == '/'
}

// SDK暂时没有跟进微信最新的AI回复，这里提供一个变通方法。
// https://developers.weixin.qq.com/doc/service/guide/product/message/Passive_user_reply_message.html#%E8%BD%AC%E6%8E%A5AI%E5%9B%9E%E5%A4%8D
type aiReply struct {
	message.CommonToken
}

func (reply *aiReply) SetMsgType(_ message.MsgType) {
	reply.MsgType = "transfer_biz_ai_ivr"
}

func (i *Ctx) handleAIReply(m *message.MixMessage) (*message.Reply, bool) {
	switch m.MsgType {
	case message.MsgTypeEvent:
		return nil, false
	case message.MsgTypeText:
		if isTextACommand(m.Content) {
			return nil, false
		}
		return &message.Reply{MsgType: message.MsgTypeTransfer, MsgData: &aiReply{}}, true
	default:
		return &message.Reply{MsgType: message.MsgTypeTransfer, MsgData: &aiReply{}}, true
	}
}
