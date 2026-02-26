package logic

import (
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	hutil "zsxyww.com/wts/handler/handlerUtilities"
	"zsxyww.com/wts/model/sqlc"
)

func wtsTextOpt(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func wtsBlockOpt(s string) sqlc.NullWtsBlock {
	if s == "" {
		return sqlc.NullWtsBlock{Valid: false}
	}
	return sqlc.NullWtsBlock{WtsBlock: sqlc.WtsBlock(s), Valid: true}
}

func wtsIspOpt(s string) sqlc.NullWtsIsp {
	if s == "" {
		return sqlc.NullWtsIsp{Valid: false}
	}
	return sqlc.NullWtsIsp{WtsIsp: sqlc.WtsIsp(s), Valid: true}
}

func timestamptzOpt(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}

}

func dateOpt(t time.Time) pgtype.Date {
	if t.IsZero() {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: t, Valid: true}
}

func wtsPriorityOpt(s string) sqlc.NullWtsPriority {
	if s == "" {
		return sqlc.NullWtsPriority{Valid: false}
	}
	return sqlc.NullWtsPriority{WtsPriority: sqlc.WtsPriority(s), Valid: true}
}

func wtsStatusOpt(s string) sqlc.NullWtsStatus {
	if s == "" {
		return sqlc.NullWtsStatus{Valid: false}
	}
	return sqlc.NullWtsStatus{WtsStatus: sqlc.WtsStatus(s), Valid: true}
}

func wtsCategoryOpt(s string) sqlc.NullWtsCategory {
	if s == "" {
		return sqlc.NullWtsCategory{Valid: false}
	}
	return sqlc.NullWtsCategory{WtsCategory: sqlc.WtsCategory(s), Valid: true}
}

func isNewStatusValid(now sqlc.WtsStatus, target sqlc.WtsStatus) bool {
	switch now {
	case sqlc.WtsStatusFresh:
		return freshTo[target]
	case sqlc.WtsStatusDelay:
		return NoPresentTo[target]
	case sqlc.WtsStatusScheduled:
		return ScheduledTo[target]
	case sqlc.WtsStatusEscalated:
		return EscalatedTo[target]
	case sqlc.WtsStatusCanceled:
		return CanceledTo[target]
	case sqlc.WtsStatusSolved:
		return SolvedTo[target]
	default:
		panic("未知的工单状态，无法判断状态转换合法性")
	}

}

var freshTo = map[sqlc.WtsStatus]bool{
	sqlc.WtsStatusFresh:     false,
	sqlc.WtsStatusDelay:     true,
	sqlc.WtsStatusScheduled: true,
	sqlc.WtsStatusEscalated: true,
	sqlc.WtsStatusCanceled:  true,
	sqlc.WtsStatusSolved:    true,
}

var NoPresentTo = map[sqlc.WtsStatus]bool{
	sqlc.WtsStatusFresh:     false,
	sqlc.WtsStatusDelay:     true,
	sqlc.WtsStatusScheduled: true,
	sqlc.WtsStatusEscalated: true,
	sqlc.WtsStatusCanceled:  true,
	sqlc.WtsStatusSolved:    true,
}

var ScheduledTo = map[sqlc.WtsStatus]bool{
	sqlc.WtsStatusFresh:     false,
	sqlc.WtsStatusDelay:     true,
	sqlc.WtsStatusScheduled: true,
	sqlc.WtsStatusEscalated: true,
	sqlc.WtsStatusCanceled:  true,
	sqlc.WtsStatusSolved:    true,
}

var EscalatedTo = map[sqlc.WtsStatus]bool{
	sqlc.WtsStatusFresh:     false,
	sqlc.WtsStatusDelay:     false,
	sqlc.WtsStatusScheduled: true,
	sqlc.WtsStatusEscalated: true,
	sqlc.WtsStatusCanceled:  true,
	sqlc.WtsStatusSolved:    true,
}

var CanceledTo = map[sqlc.WtsStatus]bool{
	sqlc.WtsStatusFresh:     false,
	sqlc.WtsStatusDelay:     false,
	sqlc.WtsStatusScheduled: false,
	sqlc.WtsStatusEscalated: false,
	sqlc.WtsStatusCanceled:  false,
	sqlc.WtsStatusSolved:    false,
}

var SolvedTo = map[sqlc.WtsStatus]bool{
	sqlc.WtsStatusFresh:     true, //可能会有误点的情况，允许管理层重新打开...
	sqlc.WtsStatusDelay:     false,
	sqlc.WtsStatusScheduled: false,
	sqlc.WtsStatusEscalated: false,
	sqlc.WtsStatusCanceled:  false,
	sqlc.WtsStatusSolved:    false,
}

// 供handler层检查权限使用
func IsOwningTicket(c *hutil.WtsCtx, user string, tidInt int32) (bool, error) {
	ctx := c.Request().Context()
	var result bool
	id := c.Response().Header().Get(echo.HeaderXRequestID)
	err := c.DB.DoQuery(ctx, "system", func(q *sqlc.Queries) error {
		a, err := q.GetTicket(ctx, tidInt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				result = false
				return hutil.NewWtsErr(ErrNoSuchTicket, err)
			}
			return err
		}
		if a.Issuer == user {
			result = true
		} else {
			result = false
		}

		return nil
	})
	if hutil.IsKnownErr(err) {
		return result, err
	} else {
		slog.Warn("IsOwningTicket数据库操作失败", "id", id, "error", err)
		if c.Cfg.Debug.APIVerbose {
			return result, err
		}
		return result, errors.New("database operation failed,please view logs")
	}
}

func emptyTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.String()
}

func emptyText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func timeOptOut(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func dateOptOut(t pgtype.Date) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func timePtrOptOut(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func datePtrOptOut(t pgtype.Date) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
