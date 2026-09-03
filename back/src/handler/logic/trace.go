package logic

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
	"zsxyww.com/wts/wechat"
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

	// 在如下情况下，向报修人发送微信通知
	if err == nil &&
		(r.NewStatus == sqlc.WtsStatusSolved ||
			r.NewStatus == sqlc.WtsStatusCanceled ||
			r.NewStatus == sqlc.WtsStatusEscalated) {
		go notifyNewStatus(c, r.Tid, r.NewStatus, r.Remark)
	}

	return err
}

// TODO: 可以直接传递数据库和微信，避免使用Echo给HTTP创建的上下文
func notifyNewStatus(c *hutil.WtsCtx, tid int32, newStatus sqlc.WtsStatus, remark string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	issuer, err := getTicketIssuer(c, ctx, tid)
	if err != nil {
		slog.Warn("新状态通知：查询工单报修人失败", "tid", tid, "error", err)
		return
	}
	openid, err := getWXBySID(c, ctx, issuer)
	if err != nil || openid == "" {
		slog.Warn("新状态通知：查询报修人OpenID失败", "tid", tid, "issuer", issuer, "error", err)
		return
	}

	var templateID string
	err = c.DB.DoQuery(ctx, "system", func(q *sqlc.Queries) error {
		var e error
		templateID, e = q.KVGet(ctx, "WX_NOTIFY_NEW_STATUS_TEMPLATE_ID")
		return e
	})
	if err != nil || templateID == "" {
		slog.Warn("新状态通知：获取模板ID时出错或未配置模板ID（kvstore: WX_NOTIFY_NEW_STATUS_TEMPLATE_ID）", "tid", tid, "error", err)
		return
	}

	// TODO:这里的关键词等是硬编码的，可以考虑改一改实现
	data := map[string]string{
		"character_string1": fmt.Sprintf("%d", tid),
		"phrase2":           statusText(newStatus),
		"thing3":            truncateRunes(remark, 20),
	}
	page := "https://wwbx.daivsye.cn/repair/"

	if err := wechat.SendNotify(c.WX, openid, templateID, data, page, true); err != nil {
		slog.Warn("新状态通知：发送失败（用户可能未订阅或额度已用完）", "tid", tid, "openid", openid, "error", err)
		return
	}
	slog.Info("新状态通知：发送成功", "tid", tid, "openid", openid, "status", newStatus)
}

// statusText 将工单状态转为通知里展示的中文
func statusText(s sqlc.WtsStatus) string {
	switch s {
	case sqlc.WtsStatusSolved:
		return "已解决"
	case sqlc.WtsStatusCanceled:
		return "已取消"
	case sqlc.WtsStatusEscalated:
		return "已上报"
	}
	return string(s)
}

// truncateRunes 按字符数截断（微信 thing 类型关键词限 20 字符）
func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

type AppendTraceParam struct {
	Tid            int32
	NewStatus      sqlc.WtsStatus
	NewPriority    sqlc.WtsPriority
	NewAppointment time.Time
	NewCategory    sqlc.WtsCategory
	Remark         string
}
