package logic

import (
	"errors"

	"github.com/jackc/pgx/v5"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

// 可能返回的错误：
// ErrNoSuchTicket: 无此工单
func GetTraces(c *hutil.WtsCtx, op string, tidInt int32) hutil.GetTracesResponse {
	ctx := c.Request().Context()

	var result hutil.GetTracesResponse

	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {
		traces, err := q.ListTracesByTicket(ctx, tidInt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) { //数据库中每个Ticket在创建的时候至少会有一条Trace记录，因此不应该出现没有记录的情况，如果没有记录说明没有这个工单
				return hutil.NewWtsErr(ErrNoSuchTicket, err)
			}
			return hutil.NewUnknownErr(err)
		}
		for _, tr := range traces {
			result.Traces = append(result.Traces, hutil.Trace{
				Opid:           tr.Opid,
				Tid:            tr.Tid,
				UpdatedAt:      timeOptOut(tr.UpdatedAt),
				Op:             tr.Op,
				OpName:         emptyText(tr.Name),
				NewStatus:      string(tr.NewStatus.WtsStatus),
				NewPriority:    string(tr.NewPriority.WtsPriority),
				NewAppointment: datePtrOptOut(tr.NewAppointment),
				NewCategory:    string(tr.NewCategory.WtsCategory),
				Remark:         tr.Remark,
				//TODO： 是不是也写emptyStatus这种的？
			})
		}
		return nil
	})

	result.Err = err
	return result

}
