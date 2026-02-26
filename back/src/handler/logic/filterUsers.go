package logic

import (
	"fmt"

	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

func FilterUsers(c *hutil.WtsCtx, op string, r hutil.FilterUsersRequest) hutil.FilterUsersResponse {
	ctx := c.Request().Context()

	var result hutil.FilterUsersResponse

	//执行数据库操作
	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {
		u, err := q.FilterUsers(ctx, sqlc.FilterUsersParams{
			Name:    wtsTextOpt(r.Name),
			Phone:   wtsTextOpt(r.Phone),
			Block:   wtsBlockOpt(r.Block),
			Room:    wtsTextOpt(r.Room),
			Isp:     wtsIspOpt(r.ISP),
			Account: wtsTextOpt(r.Account),
		})
		if err != nil {
			return hutil.NewUnknownErr(fmt.Errorf("FilterUsers::FilterUsers()出现错误: %w", err))
		}
		for _, a := range u {
			var access string
			if a.Access.WtsAccess != "" {
				access = string(a.Access.WtsAccess)
			} else {
				access = string(sqlc.WtsAccessUser)
			}
			result.Profiles = append(result.Profiles, hutil.UserProfile{
				Sid:     a.Sid.String,
				Name:    a.Name.String,
				Block:   string(a.Block.WtsBlock),
				Access:  access,
				Room:    a.Room.String,
				Phone:   a.Phone.String,
				ISP:     string(a.Isp.WtsIsp),
				Account: a.Account.String,
				WX:      a.Wx,
			})

		}
		//提交事务
		return nil
	})
	result.Err = err
	return result

}
