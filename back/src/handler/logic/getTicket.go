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
func GetTicket(c *hutil.WtsCtx, op string, who string) hutil.GetTicketResponse {

	ctx := c.Request().Context()
	result := hutil.GetTicketResponse{}

	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {
		u, err := q.GetUserByWX(ctx, who)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return hutil.NewWtsErr(ErrNoSuchUser, err)
			}
			return hutil.NewUnknownErr(fmt.Errorf("GetTicket::GetUserByWX数据库操作失败: %w", err))
		}

		var access string
		if u.Access.WtsAccess == "" {
			access = string(sqlc.WtsAccessUser)
		} else {
			access = string(u.Access.WtsAccess)
		}
		_ = access

		t, err := q.ListTicketsByIssuer(ctx, u.Sid.String)
		if err != nil {
			return hutil.NewUnknownErr(fmt.Errorf("GetTicket::ListTicketsByIssuer数据库操作失败: %w", err))
		}

		for _, ti := range t {
			result.Tickets = append(result.Tickets, hutil.Ticket{
				Tid:           ti.Tid,
				SubmittedAt:   timeOptOut(ti.SubmittedAt),
				OccurAt:       timePtrOptOut(ti.OccurAt),
				Description:   ti.Description,
				AppointedAt:   datePtrOptOut(ti.AppointedAt),
				Notes:         ti.Notes.String,
				Priority:      string(ti.Priority),
				Category:      string(ti.Category),
				Status:        string(ti.Status),
				LastUpdatedAt: timeOptOut(ti.LastUpdatedAt),
				Issuer: hutil.UserProfile{
					Sid:     ti.Issuer,
					Name:    ti.Name.String,
					Block:   string(ti.Block.WtsBlock),
					Access:  access,
					Room:    ti.Room.String,
					Phone:   ti.Phone.String,
					ISP:     string(ti.Isp.WtsIsp),
					Account: ti.Account.String,
				},
			})
		}

		return nil
	})

	result.Err = err
	return result

}
