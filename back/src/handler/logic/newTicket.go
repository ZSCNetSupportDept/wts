package logic

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

// 可能返回的错误：
// ErrNoSuchUser: 无此用户
// ErrAppointTimeInvalid: 预约时间无效（早于当前时间）
// ErrOccurAtTimeInvalid: 发生时间无效（晚于当前时间）
// ErrDataInconsistent: 数据库返回数据前后不一致
// ErrTicketTooMuch: 用户未结工单过多
func NewTicket(c *hutil.WtsCtx, op string, r hutil.NewTicketRequest) hutil.NewTicketResponse {
	ctx := c.Request().Context()
	result := hutil.NewTicketResponse{}

	err := c.DB.DoQuery(ctx, op, func(q *sqlc.Queries) error {

		var err error

		// 检查报修提交的时间是否在值班时间内
		var dutyStart, dutyEnd string
		var dutyStartTime, dutyEndTime time.Time

		if r.AppointedAt.IsZero() {
			goto dutyCheckFinish
		}

		dutyStart, err = q.KVGet(ctx, "WTS_DUTY_START")
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.Warn("检测到未设置WTS_DUTY_START，跳过值班时间检查")
				goto dutyCheckFinish
			} else {
				return hutil.NewUnknownErr(fmt.Errorf("NewTicket::KVGet数据库操作失败: %w", err))
			}
		}
		dutyEnd, err = q.KVGet(ctx, "WTS_DUTY_END")
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				slog.Warn("检测到未设置WTS_DUTY_END，跳过值班时间检查")
				goto dutyCheckFinish
			} else {
				return hutil.NewUnknownErr(fmt.Errorf("NewTicket::KVGet数据库操作失败: %w", err))
			}
		}

		dutyStartTime, err = time.Parse(time.RFC3339, dutyStart)
		if err != nil {
			slog.Warn("WTS_DUTY_START不符合RFC3339格式，跳过值班时间检查，")
			goto dutyCheckFinish
		}
		dutyEndTime, err = time.Parse(time.RFC3339, dutyEnd)
		if err != nil {
			slog.Warn("WTS_DUTY_END不符合RFC3339格式，跳过值班时间检查")
			goto dutyCheckFinish
		}
		if dutyStartTime.After(dutyEndTime) {
			slog.Warn("WTS_DUTY_START晚于WTS_DUTY_END，跳过值班时间检查")
			goto dutyCheckFinish
		}
		if r.AppointedAt.Before(dutyStartTime) {
			return hutil.NewWtsErr(errors.Join(ErrBeforeDutyStart, errors.New("我们开始值班的时间为"+dutyStartTime.Format(time.DateOnly))), nil)
		}
		if r.AppointedAt.After(dutyEndTime) {
			return hutil.NewWtsErr(errors.Join(ErrAfterDutyEnd, errors.New("我们结束值班的时间为"+dutyEndTime.Format(time.DateOnly))), nil)
		}

	dutyCheckFinish:
		//检查时间是否是合理的
		if !r.AppointedAt.IsZero() && r.AppointedAt.Before(time.Now()) { //TODO:预约可以在今天！！！重要！！！
			return hutil.NewWtsErr(ErrAppointTimeInvalid, nil)
		}

		if !r.OccurAt.IsZero() && r.OccurAt.After(time.Now()) {
			return hutil.NewWtsErr(ErrOccurAtTimeInvalid, nil)
		}

		//获取报修人信息
		u, err := q.GetUserBySID(ctx, wtsTextOpt(r.IssuerSID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return hutil.NewWtsErr(ErrNoSuchUser, err)
			}
			return hutil.NewUnknownErr(fmt.Errorf("NewTicket::GetUserBySID数据库操作失败: %w", err))
		}

		i := u.Sid.String
		//检查用户是否有过多未结工单
		//fmt.Println("用户SID:", u.Sid)
		count := 0
		a, err := q.ListTicketsByIssuer(ctx, u.Sid.String)
		if err != nil {
			return hutil.NewUnknownErr(fmt.Errorf("GetTicket::ListTicketsByIssuer数据库操作失败: %w", err))
		}
		for _, ti := range a {
			if !(ti.Status == sqlc.WtsStatusCanceled || ti.Status == sqlc.WtsStatusSolved) {
				count++
			}
		}
		//fmt.Println("未结工单数量:", count)
		if count >= 1 {
			return hutil.NewWtsErr(ErrTicketTooMuch, nil)
		}

		//走到这里应该没问题，那创建工单
		t, err := q.CreateTicket(ctx, sqlc.CreateTicketParams{
			Issuer:      i,
			SubmittedAt: timestamptzOpt(time.Now()),
			OccurAt:     timestamptzOpt(r.OccurAt),
			Description: r.Description,
			AppointedAt: dateOpt(r.AppointedAt),
			Notes:       wtsTextOpt(r.Notes),
			Priority:    sqlc.WtsPriority(r.Priority),
			Category:    sqlc.WtsCategory(r.Category),
			Status:      sqlc.WtsStatus(r.Status),
		})
		if err != nil {
			return hutil.NewUnknownErr(fmt.Errorf("NewTicket::CreateTicket数据库操作失败: %w", err))
		}

		if (t.Issuer != i) &&
			(t.Description != r.Description) &&
			(t.Category != sqlc.WtsCategory(r.Category)) &&
			(t.Priority != sqlc.WtsPriority(r.Priority)) &&
			(t.Status != sqlc.WtsStatus(r.Status)) &&
			(t.Notes.String == r.Notes) &&
			(t.OccurAt.Time.Equal(r.OccurAt)) &&
			((t.AppointedAt.Time.Equal(r.AppointedAt)) || (r.AppointedAt.IsZero())) {
			return hutil.NewWtsErr(ErrDataInconsistent, nil)
		}

		result.Tid = t.Tid

		//提交事务
		return nil
	})
	result.Err = err
	return result
}
