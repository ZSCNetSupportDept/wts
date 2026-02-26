package logic

import (
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
		return "同学你好，欢迎使用网维报修系统~\n\n建议先看看使用攻略呢：https://wts.zsxyww.com/self-service/usage\n"
	}
	if m.Event == message.EventView {
		// 不知道为什么，view事件也会被送到这里，如果不处理的话会在log里面出现，有点烦，对用户体验倒是没什么影响
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
	return "聊天功能正在开发中"
	//return i.superEasyNLPProgram(m)
}

func (i *Ctx) superEasyNLPProgram(m *message.MixMessage) string {
	verb := string([]rune(m.Content)[0])
	return verb + verb + verb + "，一天到晚就™知道" + m.Content + "，是不是军姿没站够？"
}
