package logic

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

// 可能返回的错误：
// ErrNoSuchUser: 无此用户
func ViewProfile(c *hutil.WtsCtx, op string, who string) hutil.ViewUserProfileResponse {

	ctx := c.Request().Context()
	result := hutil.ViewUserProfileResponse{}

	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {
		user, err := q.GetUserByWX(ctx, who)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return hutil.NewWtsErr(ErrNoSuchUser, err)
			}
			return hutil.NewUnknownErr(fmt.Errorf("ViewProfile::GetUserByWX数据库操作失败: %w", err))
		}
		var access string
		if user.Access.WtsAccess != "" {
			access = string(user.Access.WtsAccess)
		} else {
			access = string(sqlc.WtsAccessUser)
		}
		result.Profile = hutil.UserProfile{
			Sid:     user.Sid.String,
			Name:    user.Name.String,
			Block:   string(user.Block.WtsBlock),
			Access:  access,
			Room:    user.Room.String,
			Phone:   user.Phone.String,
			ISP:     string(user.Isp.WtsIsp),
			Account: user.Account.String,
			WX:      user.Wx,
		}

		return nil
	})

	result.Err = err
	return result

}
