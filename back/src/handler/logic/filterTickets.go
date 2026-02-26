package logic

import (
	"fmt"

	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

// 可能返回的错误：
// ErrInvalidScope: 无效的Scope参数
// ErrInvalidZone: 无效的片区参数
func FilterTickets(c *hutil.WtsCtx, op string, r hutil.FilterTicketsRequest) hutil.FilterTicketsResponse {
	ctx := c.Request().Context()

	var result hutil.FilterTicketsResponse
	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {
		var err error
		var t []sqlc.WtsVTicket
		//执行数据库操作
		switch r.Scope {
		case "active":
			t, err = wrap(q.FilterActiveTickets(ctx, sqlc.FilterActiveTicketsParams{
				Blocks:    r.Block,
				Issuer:    wtsTextOpt(r.Issuer),
				Category:  r.Category,
				Isp:       r.ISP,
				NewerThan: timestamptzOpt(r.NewerThan),
				OlderThan: timestamptzOpt(r.OlderThan),
				Status:    r.Status,
			}))
			if err != nil {
				return hutil.NewUnknownErr(fmt.Errorf("FilterTickets::FilterTickets()出现错误: %w", err))
			}
		case "all":
			t, err = q.FilterTickets(ctx, sqlc.FilterTicketsParams{
				Blocks:    r.Block,
				Issuer:    wtsTextOpt(r.Issuer),
				Category:  r.Category,
				Isp:       r.ISP,
				NewerThan: timestamptzOpt(r.NewerThan),
				OlderThan: timestamptzOpt(r.OlderThan),
				Status:    r.Status,
			})
			if err != nil {
				return hutil.NewUnknownErr(fmt.Errorf("FilterTickets::FilterTickets()出现错误: %w", err))
			}
		default:
			return hutil.NewWtsErr(ErrInvalidScope, nil)

		}

		for _, a := range t {
			result.Tickets = append(result.Tickets, hutil.Ticket{
				Tid: a.Tid,
				Issuer: hutil.UserProfile{
					Sid:     a.Issuer,
					Name:    a.Name.String,
					Block:   string(a.Block.WtsBlock),
					Room:    a.Room.String,
					Phone:   a.Phone.String,
					ISP:     string(a.Isp.WtsIsp),
					Account: a.Account.String,
				},
				SubmittedAt:   timeOptOut(a.SubmittedAt),
				OccurAt:       timePtrOptOut(a.OccurAt),
				Description:   a.Description,
				Category:      string(a.Category),
				Notes:         a.Notes.String,
				Priority:      string(a.Priority),
				Status:        string(a.Status),
				AppointedAt:   datePtrOptOut(a.AppointedAt),
				LastUpdatedAt: timeOptOut(a.LastUpdatedAt),
			})
		}

		return nil
	})
	result.Err = err
	return result
}

func wrap(t []sqlc.WtsVActiveTicket, e error) ([]sqlc.WtsVTicket, error) {
	var res []sqlc.WtsVTicket
	for _, v := range t {
		res = append(res, sqlc.WtsVTicket(v))
	}
	return res, e
}
