package logic

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"github.com/silenceper/wechat/v2/util"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

func (i *Ctx) processCommand(m *message.MixMessage) string {
	//可以假定进入这里的信息全部是以"/"开头的Text类型

	//user := i.WX.GetUser()
	u := m.GetOpenID()

	switch m.Content {
	case "/debug OpenID":
		return "您的OpenID是:" + m.GetOpenID()
	case "/tagme default":
		err := i.ChangeUserTag(u, "default")
		if err != nil {
			return "操作失败：" + err.Error()
		}
		return "您的标签已设置为默认用户~"
	case "/tagme operator":
		who, err1 := getUserByWX((*hutil.WtsCtx)(i), u)
		if err1 != nil {
			return "操作失败：" + err1.Error()
		}
		if !hutil.IsOperator(sqlc.WtsAccess(who.Access)) {
			return "您还不是网维的成员，无法进行此操作~"
		}
		err := i.ChangeUserTag(u, "operator")
		if err != nil {
			return "操作失败：" + err.Error()
		}
		return "您的标签已设置为网维成员~"
	case "/tagme admin":
		who, err1 := getUserByWX((*hutil.WtsCtx)(i), u)
		if err1 != nil {
			return "操作失败：" + err1.Error()
		}
		if !hutil.IsAdmin(sqlc.WtsAccess(who.Access)) {
			return "您还不是管理层，无法进行此操作~"
		}
		err := i.ChangeUserTag(u, "admin")
		if err != nil {
			return "操作失败：" + err.Error()
		}
		return "您的标签已设置为管理层成员~"
	case "/auth":
		return i.auth(m.GetOpenID())
	case "/deauth":
		err := i.ChangeUserTag(u, "default") //把菜单栏改回用户的微信菜单栏
		if err != nil {
			return "操作失败：" + err.Error()
		}
		return "菜单已改回默认用户菜单，若改回则重新输入/auth即可~"
	default:
		return "无法识别的命令，请重新输入~"
	}
}

func (i *Ctx) ChangeUserTag(oid string, tag string) error {
	u := i.WX.GetUser()

	p := []string{oid}
	var err error

	checkUntagErr := func(err error) error {
		if err != nil {
			var commonError *util.CommonError
			if errors.As(err, &commonError) {
				// 微信错误码 45059 表示 "用户未打上该标签"，我们忽略它
				if commonError.ErrCode == 45059 {
					return nil
				}
			}
			return err
		}
		return nil
	}

	switch tag {
	case "default":
		err = checkUntagErr(u.BatchUntag(p, 100))
		if err != nil {
			return err
		}
		err = checkUntagErr(u.BatchUntag(p, 101))
		if err != nil {
			return err
		}
	case "operator":
		err = checkUntagErr(u.BatchUntag(p, 101))
		if err != nil {
			return err
		}
		err = u.BatchTag(p, 100)
		if err != nil {
			return err
		}
	case "admin":
		err = checkUntagErr(u.BatchUntag(p, 100))
		if err != nil {
			return err
		}
		err = u.BatchTag(p, 101)
		if err != nil {
			return err
		}
	default:
		return errors.New("unknown tag: " + tag)

	}
	return nil

}

// 检查用户是否已经在operators或admin表中，若是则修改菜单栏为对应权限的。
func (i *Ctx) auth(u string) string {
	ctx := context.Background()
	err := i.DB.DoQuery(ctx, u, func(q *sqlc.Queries) error {
		u, err := q.GetUserByWX(ctx, u)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("您还未绑定，无法进行此操作~")
			}
			return errors.New("数据库错误：" + err.Error())
		}
		if !u.Op {
			return errors.New("您还不是网维的成员，无法进行此操作~")
		}
		if hutil.IsAdmin(u.Access.WtsAccess) {
			i.ChangeUserTag(u.Wx, "admin")
			return nil
		}
		i.ChangeUserTag(u.Wx, "operator")
		return nil
	})
	if err != nil {
		return "操作失败：" + err.Error()
	}
	return "认证成功，您的微信菜单已更新，请取关重关注公众号以刷新菜单~"
}
