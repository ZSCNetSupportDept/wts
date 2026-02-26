package logic

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

// 可能返回的错误：
// ErrNoSuchTicket: 无此工单
// ErrNoSuchStaff: 无此网维成员
// ErrNewStatusInvalid: 工单新状态不符合逻辑
// ErrDataInconsistent: 数据库返回数据前后不一致
func AppendTrace(c *hutil.WtsCtx, op string, r AppendTraceParam) error {

	ctx := c.Request().Context()

	var opwid string

	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {

		//确认工单存在
		t, err := q.GetTicket(ctx, r.Tid)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return hutil.NewWtsErr(ErrNoSuchTicket, err)
			}
			return hutil.NewUnknownErr(fmt.Errorf("AppendTrace::GetTicket数据库操作失败: %w", err))
		}

		//确认记录添加人的信息和有效性
		w, err := q.GetUserByWX(ctx, op)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return hutil.NewWtsErr(ErrNoSuchStaff, err)
			}
			return hutil.NewUnknownErr(fmt.Errorf("AppendTrace::GetUserByWX数据库操作失败: %w", err))
		}
		if !w.Op {
			opwid = "-2"
		} else {
			opw, err := q.GetStaffBySid(ctx, w.Sid.String)
			if err != nil {
				return hutil.NewUnknownErr(fmt.Errorf("AppendTrace::GetStaffBySid数据库操作失败: %w", err))
			}
			opwid = opw.Wid
		}

		//确认新状态是合乎逻辑的...
		if r.NewStatus != "" {
			if !isNewStatusValid(t.Status, r.NewStatus) {
				return hutil.NewWtsErr(ErrNewStatusInvalid, nil)
			}
		}

		tr, err := q.CreateTicketTrace(ctx, sqlc.CreateTicketTraceParams{
			Tid:            r.Tid,
			UpdatedAt:      timestamptzOpt(time.Now()),
			Op:             opwid,
			NewStatus:      wtsStatusOpt(string(r.NewStatus)),
			NewPriority:    wtsPriorityOpt(string(r.NewPriority)),
			NewAppointment: dateOpt(r.NewAppointment),
			NewCategory:    wtsCategoryOpt(string(r.NewCategory)),
			Remark:         r.Remark,
		})
		if err != nil {
			return hutil.NewUnknownErr(fmt.Errorf("AppendTrace::CreateTicketTrace数据库操作失败: %w", err))
		}
		if !((tr.Tid == r.Tid) &&
			(tr.Op == opwid) &&
			(tr.Remark == r.Remark) &&
			((r.NewStatus == "" && !tr.NewStatus.Valid) || (tr.NewStatus.WtsStatus == r.NewStatus)) &&
			((r.NewPriority == "" && !tr.NewPriority.Valid) || (tr.NewPriority.WtsPriority == r.NewPriority)) &&
			((r.NewCategory == "" && !tr.NewCategory.Valid) || (tr.NewCategory.WtsCategory == r.NewCategory)) &&
			((r.NewAppointment.IsZero() && !tr.NewAppointment.Valid) || (tr.NewAppointment.Time.Format(time.DateOnly) == r.NewAppointment.Format(time.DateOnly)))) {
			return hutil.NewWtsErr(ErrDataInconsistent, nil)
		}

		return nil

	})
	return err
}

type AppendTraceParam struct {
	Tid            int32
	NewStatus      sqlc.WtsStatus
	NewPriority    sqlc.WtsPriority
	NewAppointment time.Time
	NewCategory    sqlc.WtsCategory
	Remark         string
}
