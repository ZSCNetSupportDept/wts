package logic

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	. "zsxyww.com/wts/handler/handlerUtilities"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

// 可能返回的错误：
// ErrNoSuchUser: 无此用户
// ErrPhoneUsed: 电话号码已被使用
// ErrDataInconsistent: 数据库操作后数据不一致
func ChangeProfile(c *WtsCtx, op string, who string, r ChangeUserProfileRequest) ChangeUserProfileResponse {
	ctx := c.Request().Context()
	result := ChangeUserProfileResponse{}

	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {
		user, err := q.GetUserByWX(ctx, who)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return NewWtsErr(ErrNoSuchUser, err)
			}
			return NewUnknownErr(fmt.Errorf("ChangeProfile::GetUserByWX数据库操作失败: %w", err))
		}

		db, err := q.UpdateUser(ctx, sqlc.UpdateUserParams{
			Sid:     user.Sid.String,
			Block:   sqlc.NullWtsBlock{WtsBlock: sqlc.WtsBlock(r.Block), Valid: (r.Block != "")},
			Room:    pgtype.Text{String: r.Room, Valid: (r.Room != "")},
			Phone:   pgtype.Text{String: r.Phone, Valid: (r.Phone != "")},
			Isp:     sqlc.NullWtsIsp{WtsIsp: sqlc.WtsIsp(r.ISP), Valid: (r.ISP != "")},
			Account: pgtype.Text{String: r.Account, Valid: (r.Account != "")},
		})
		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "phone_unique":
					return NewWtsErr(ErrPhoneUsed, err)
				}
				return NewUnknownErr(fmt.Errorf("ChangeProfile::UpdateUser数据库操作失败: %w", err))
			}
		}

		// 检查创建结果是否和请求一致
		i := (user.Sid.String == db.Sid) &&
			(db.Wx == who) && (user.Wx == who)
		if !i {
			return hutil.NewWtsErr(ErrDataInconsistent, nil)
		}

		//提交事务
		return nil
	})

	result.Err = err
	return result

}
