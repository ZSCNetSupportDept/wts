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
// ErrSidNameNotMatch: 学号与姓名不匹配
// ErrNoStudentRecord: 无此学生记录
// ErrUserAlreadyRegistered: 用户已注册
// ErrPhoneUsed: 电话号码已被使用
// ErrWxUsed: 该微信号已被使用
// ErrDataInconsistent: 数据库操作后数据不一致
func Register(c *WtsCtx, op string, r hutil.RegisterRequest) hutil.RegisterResponse {

	//执行数据库操作
	//TODO:优化错误处理与返回，增加日志模块,文档...
	ctx := c.Request().Context()
	wx := op

	result := hutil.RegisterResponse{}

	err := c.DB.DoQuery(ctx, wx, func(q *sqlc.Queries) error {
		//开始事务

		//检查用户发来的姓名和学号是否对应
		name, err := q.GetNameBySID(ctx, r.Sid)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return hutil.NewWtsErr(ErrNoStudentRecord, err)
			}
			return hutil.NewUnknownErr(fmt.Errorf("Register::GetNameBySID数据库操作失败: %w", err))
		}
		if name != r.Name {
			return hutil.NewWtsErr(ErrSidNameNotMatch, err)
		}

		// 如果能走到这里，说明数据没有问题，那么，进行实际的创建用户操作
		// 创建用户
		db, err := q.CreateUser(ctx, sqlc.CreateUserParams{
			Sid:     r.Sid,
			Block:   sqlc.NullWtsBlock{WtsBlock: sqlc.WtsBlock(r.Block), Valid: (r.Block != "")},
			Room:    pgtype.Text{String: r.Room, Valid: (r.Room != "")},
			Phone:   pgtype.Text{String: r.Phone, Valid: (r.Phone != "")},
			Isp:     sqlc.NullWtsIsp{WtsIsp: sqlc.WtsIsp(r.ISP), Valid: (r.ISP != "")},
			Account: pgtype.Text{String: r.Account, Valid: (r.Account != "")},
			Wx:      wx,
		})

		if err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				switch pgErr.ConstraintName {
				case "users_pk":
					return hutil.NewWtsErr(ErrUserAlreadyRegistered, err)
				case "phone_unique":
					return hutil.NewWtsErr(ErrPhoneUsed, err)
				case "idx_wx_unique":
					return hutil.NewWtsErr(ErrWxUsed, err)
				}
			}
			return hutil.NewUnknownErr(fmt.Errorf("Register::CreateUser数据库操作失败: %w", err))
		}

		// 检查创建结果是否和请求一致
		i := (r.Sid == db.Sid) &&
			(db.Wx == wx) &&
			(db.Block.WtsBlock == sqlc.WtsBlock(r.Block)) &&
			(db.Room.String == r.Room) &&
			(db.Phone.String == r.Phone) &&
			(db.Isp.WtsIsp == sqlc.WtsIsp(r.ISP)) &&
			(db.Account.String == r.Account)
		if !i {
			return hutil.NewWtsErr(ErrDataInconsistent, nil)
		}

		//提交事务
		return nil
	})

	result.Err = err
	return result

}
