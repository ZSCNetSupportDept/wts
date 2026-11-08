package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Swap struct {
	ID            int64  `json:"id"`
	RequesterWid  string `json:"requester_wid"`
	RequesterName string `json:"requester_name"`
	TargetWid     string `json:"target_wid"`
	TargetName    string `json:"target_name"`
	DateA         string `json:"date_a"`
	DateB         string `json:"date_b"`
	RevisionA     int    `json:"revision_a"`
	RevisionB     int    `json:"revision_b"`
	ApprovalA     string `json:"approval_a"`
	ApprovalB     string `json:"approval_b"`
	Status        string `json:"status"`
	Reason        string `json:"reason"`
	CanReview     bool   `json:"can_review"`
	CanClaim      bool   `json:"can_claim"`
}
type SwapRequest struct {
	DateA  string `json:"date_a"`
	DateB  string `json:"date_b"`
	Reason string `json:"reason"`
	ID     int64  `json:"id"`
	Action string `json:"action"`
}

const swapSelect = `SELECT r.id,r.requester_wid,COALESCE(sa.name,''),COALESCE(r.target_wid,''),COALESCE(sb.name,''),
    to_char(r.date_a,'YYYY-MM-DD'),to_char(r.date_b,'YYYY-MM-DD'),r.revision_a,r.revision_b,
    COALESCE(r.approval_a,''),COALESCE(r.approval_b,''),r.status,r.reason
    FROM scheduler.swap_requests r
    JOIN wts.operators oa ON oa.wid=r.requester_wid LEFT JOIN data.students sa ON sa.sid=oa.sid
    LEFT JOIN wts.operators ob ON ob.wid=r.target_wid LEFT JOIN data.students sb ON sb.sid=ob.sid `

func scanSwap(row pgx.Row) (Swap, error) {
	var r Swap
	err := row.Scan(&r.ID, &r.RequesterWid, &r.RequesterName, &r.TargetWid, &r.TargetName, &r.DateA, &r.DateB, &r.RevisionA, &r.RevisionB, &r.ApprovalA, &r.ApprovalB, &r.Status, &r.Reason)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, fail(404, "换班申请不存在")
	}
	return r, err
}
func (w *work) twoDays(a, b string) (Day, Day, error) {
	dates := []string{a, b}
	sort.Strings(dates)
	for _, date := range dates {
		if err := w.lockDate(date); err != nil {
			return Day{}, Day{}, err
		}
	}
	da, err := w.day(a, true)
	if err != nil {
		return da, Day{}, err
	}
	db, err := w.day(b, true)
	return da, db, err
}
func regularMember(d Day, wid string) (Member, bool) {
	for _, m := range d.Members {
		if m.Wid == wid && !m.Extra {
			return m, true
		}
	}
	return Member{}, false
}
func memberExists(d Day, wid string) bool {
	for _, m := range d.Members {
		if m.Wid == wid {
			return true
		}
	}
	return false
}
func validateSwap(a, b Day, requester, target string) error {
	if a.Date == b.Date || requester == target {
		return fail(400, "请选择不同成员和不同值班日期")
	}
	if !a.Published || !b.Published {
		return fail(400, "只能交换已发布的班次")
	}
	if _, ok := regularMember(a, requester); !ok {
		return fail(400, "申请人不在原日期的正式班次中")
	}
	if _, ok := regularMember(b, target); !ok {
		return fail(400, "对方不在目标日期的正式班次中")
	}
	if a.Leader.Wid == requester || b.Leader.Wid == target {
		return fail(400, "值班组长不能通过成员换班更换，需科长/dev/api重新配置小组")
	}
	if memberExists(a, target) || memberExists(b, requester) {
		return fail(409, "交换后会产生同一天重复值班")
	}
	today := time.Now().In(china).Format(time.DateOnly)
	if a.Date < today || b.Date < today {
		return fail(400, "不能交换已经过去的班次")
	}
	return nil
}

func validateOpenSwap(a, b Day, requester string) error {
	if a.Date == b.Date {
		return fail(400, "预期换班日期不能与原日期相同")
	}
	if !a.Published || !b.Published {
		return fail(400, "请选择已发布班次的预期日期")
	}
	if _, ok := regularMember(a, requester); !ok {
		return fail(400, "你不在原日期的正式班次中")
	}
	if a.Leader.Wid == requester {
		return fail(400, "值班组长的更换需要科长/dev/api重新配置")
	}
	if memberExists(b, requester) {
		return fail(409, "你已在预期日期值班")
	}
	today := time.Now().In(china).Format(time.DateOnly)
	if a.Date < today || b.Date < today {
		return fail(400, "不能申请已经过去的班次")
	}
	return nil
}

func (s Service) Swaps(ctx context.Context, openID, month string) (any, error) {
	start, err := time.ParseInLocation("2006-01", month, china)
	if err != nil || start.Format("2006-01") != month {
		return nil, fail(400, "月份必须为 YYYY-MM")
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		rows, err := w.tx.Query(ctx, swapSelect+` JOIN scheduler.duty_days da ON da.duty_date=r.date_a
            JOIN scheduler.duty_days db ON db.duty_date=r.date_b
            WHERE ((r.date_a >= $1::date AND r.date_a < $2::date) OR (r.date_b >= $1::date AND r.date_b < $2::date))
              AND ($3 OR r.requester_wid=$4 OR r.target_wid=$4 OR ($5 AND (da.leader_wid=$4 OR db.leader_wid=$4))
                OR (r.status='open' AND EXISTS(SELECT 1 FROM scheduler.duty_members m
                     WHERE m.duty_date=r.date_b AND m.wid=$4 AND NOT m.extra)))
            ORDER BY r.id DESC`, start.Format(time.DateOnly), start.AddDate(0, 1, 0).Format(time.DateOnly), isManager(w.actor.Access), w.actor.Wid, w.actor.Access == "group-leader")
		if err != nil {
			return nil, err
		}
		result := []Swap{}
		for rows.Next() {
			r, e := scanSwap(rows)
			if e != nil {
				rows.Close()
				return nil, e
			}
			result = append(result, r)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		for i, r := range result {
			a, e := w.day(r.DateA, false)
			if e != nil {
				return nil, e
			}
			b, e := w.day(r.DateB, false)
			if e != nil {
				return nil, e
			}
			result[i].CanReview = r.Status == "pending" && w.actor.Access == "group-leader" &&
				(a.Leader.Wid == w.actor.Wid && r.ApprovalA == "" || b.Leader.Wid == w.actor.Wid && r.ApprovalB == "")
			result[i].CanClaim = r.Status == "open" && r.RevisionA == a.Revision && r.RevisionB == b.Revision &&
				validateSwap(a, b, r.RequesterWid, w.actor.Wid) == nil
		}
		return result, nil
	})
}

func (s Service) CreateSwap(ctx context.Context, openID string, r SwapRequest) (any, error) {
	if _, err := parseDate(r.DateA); err != nil {
		return nil, err
	}
	if _, err := parseDate(r.DateB); err != nil {
		return nil, err
	}
	if len(r.Reason) > 1000 {
		return nil, fail(400, "原因过长")
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		a, b, err := w.twoDays(r.DateA, r.DateB)
		if err != nil {
			return nil, err
		}
		if err = validateOpenSwap(a, b, w.actor.Wid); err != nil {
			return nil, err
		}
		var busy bool
		err = w.tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM scheduler.swap_requests WHERE status IN ('open','pending') AND
            ((date_a=$1::date AND requester_wid=$2) OR (date_b=$1::date AND target_wid=$2)))`, a.Date, w.actor.Wid).Scan(&busy)
		if err != nil {
			return nil, err
		}
		if busy {
			return nil, fail(409, "相关班次已有待确认的换班申请")
		}
		var id int64
		err = w.tx.QueryRow(ctx, `INSERT INTO scheduler.swap_requests(requester_wid,date_a,date_b,revision_a,revision_b,reason)
            VALUES($1,$2::date,$3::date,$4,$5,$6) RETURNING id`, w.actor.Wid, a.Date, b.Date, a.Revision, b.Revision, strings.TrimSpace(r.Reason)).Scan(&id)
		if err != nil {
			return nil, err
		}
		result, err := scanSwap(w.tx.QueryRow(ctx, swapSelect+` WHERE r.id=$1`, id))
		if err != nil {
			return nil, err
		}
		return result, w.swapAudit(result, "swap_request")
	})
}
func (w *work) swapAudit(r Swap, action string) error {
	state, err := json.Marshal(r)
	if err != nil {
		return err
	}
	_, err = w.tx.Exec(w.ctx, `INSERT INTO scheduler.audit_log(actor_wid,action,reason,after_state) VALUES($1,$2,$3,$4::jsonb)`, w.actor.Wid, action, r.Reason, string(state))
	return err
}

func (s Service) ReviewSwap(ctx context.Context, openID string, r SwapRequest) (any, error) {
	if r.ID <= 0 {
		return nil, fail(400, "无效换班申请编号")
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		initial, err := scanSwap(w.tx.QueryRow(ctx, swapSelect+` WHERE r.id=$1`, r.ID))
		if err != nil {
			return nil, err
		}
		// Lock dates first, consistently with schedule edits, then the request row.
		a, b, err := w.twoDays(initial.DateA, initial.DateB)
		if err != nil {
			return nil, err
		}
		swap, err := scanSwap(w.tx.QueryRow(ctx, swapSelect+` WHERE r.id=$1 FOR UPDATE OF r`, r.ID))
		if err != nil {
			return nil, err
		}
		if swap.Status != "pending" && swap.Status != "open" {
			return nil, fail(409, "申请已处理或原排班已变化，请刷新")
		}
		ownA := w.actor.Access == "group-leader" && a.Leader.Wid == w.actor.Wid
		ownB := w.actor.Access == "group-leader" && b.Leader.Wid == w.actor.Wid
		switch r.Action {
		case "claim":
			if swap.Status != "open" {
				return nil, fail(409, "请求已被其他成员接下")
			}
			if swap.RevisionA != a.Revision || swap.RevisionB != b.Revision {
				return nil, fail(409, "原排班已变化，请刷新")
			}
			if _, ok := regularMember(b, w.actor.Wid); !ok {
				return nil, fail(403, "只有预期日期的正式班次组员可以接单")
			}
			if err = validateSwap(a, b, swap.RequesterWid, w.actor.Wid); err != nil {
				return nil, err
			}
			if _, err = w.person(swap.RequesterWid); err != nil {
				return nil, err
			}
			var busy bool
			err = w.tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM scheduler.swap_requests WHERE id<>$1 AND status IN ('open','pending')
				AND ((date_a=$2::date AND requester_wid=$3) OR (date_b=$2::date AND target_wid=$3)))`, swap.ID, b.Date, w.actor.Wid).Scan(&busy)
			if err != nil {
				return nil, err
			}
			if busy {
				return nil, fail(409, "你的班次已有未完成的换班申请")
			}
			swap.TargetWid = w.actor.Wid
			swap.TargetName = w.actor.Name
			swap.Status = "pending"
		case "cancel":
			if w.actor.Wid != swap.RequesterWid {
				return nil, fail(403, "只有申请人能撤销申请")
			}
			swap.Status = "canceled"
		case "reject":
			if swap.Status != "pending" {
				return nil, fail(409, "需要先由对应日期成员接单，再交组长确认")
			}
			if !ownA && !ownB {
				return nil, fail(403, "只有双方对应日期的组长可以确认或拒绝")
			}
			swap.Status = "rejected"
		case "approve":
			if swap.Status != "pending" {
				return nil, fail(409, "需要先由对应日期成员接单，再交组长确认")
			}
			if !ownA && !ownB {
				return nil, fail(403, "只有双方对应日期的组长可以确认或拒绝")
			}
			if swap.RevisionA != a.Revision || swap.RevisionB != b.Revision {
				return nil, fail(409, "原排班已变化，请重新申请")
			}
			if err = validateSwap(a, b, swap.RequesterWid, swap.TargetWid); err != nil {
				return nil, err
			}
			if _, err = w.person(swap.RequesterWid); err != nil {
				return nil, err
			}
			if _, err = w.person(swap.TargetWid); err != nil {
				return nil, err
			}
			if ownA {
				swap.ApprovalA = w.actor.Wid
			}
			if ownB {
				swap.ApprovalB = w.actor.Wid
			}
			if swap.ApprovalA != "" && swap.ApprovalB != "" {
				swap.Status = "approved"
			}
		default:
			return nil, fail(400, "无效确认操作")
		}
		if _, err = w.tx.Exec(ctx, `UPDATE scheduler.swap_requests SET approval_a=NULLIF($2,''),approval_b=NULLIF($3,''),status=$4,target_wid=NULLIF($5,'') WHERE id=$1`, swap.ID, swap.ApprovalA, swap.ApprovalB, swap.Status, swap.TargetWid); err != nil {
			return nil, err
		}
		if swap.Status == "approved" {
			ma, _ := regularMember(a, swap.RequesterWid)
			mb, _ := regularMember(b, swap.TargetWid)
			// Exchange the two dated slots; no weekly availability or other dates are changed.
			if _, err = w.tx.Exec(ctx, `DELETE FROM scheduler.duty_members WHERE (duty_date=$1::date AND wid=$3) OR (duty_date=$2::date AND wid=$4)`, a.Date, b.Date, ma.Wid, mb.Wid); err != nil {
				return nil, err
			}
			if _, err = w.tx.Exec(ctx, `INSERT INTO scheduler.duty_members(duty_date,wid,zone,area_leader) VALUES($1::date,$2,$3,$4),($5::date,$6,$7,$8)`, a.Date, mb.Wid, ma.Zone, ma.AreaLeader, b.Date, ma.Wid, mb.Zone, mb.AreaLeader); err != nil {
				return nil, err
			}
			if _, err = w.record(a, "swap_complete", swap.Reason); err != nil {
				return nil, err
			}
			if _, err = w.record(b, "swap_complete", swap.Reason); err != nil {
				return nil, err
			}
		}
		return swap, w.swapAudit(swap, "swap_"+r.Action)
	})
}

func (s Service) Audit(ctx context.Context, openID, date string) (any, error) {
	if _, err := parseDate(date); err != nil {
		return nil, err
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		d, err := w.day(date, false)
		if err != nil {
			return nil, err
		}
		if !canArrange(w.actor, d) {
			return nil, fail(403, "只能查看自己负责日期的操作记录")
		}
		rows, err := w.tx.Query(ctx, `SELECT jsonb_build_object('actor_wid',actor_wid,'action',action,'reason',reason,'created_at',created_at,'before_state',before_state,'after_state',after_state)
            FROM scheduler.audit_log WHERE duty_date=$1::date ORDER BY id DESC LIMIT 100`, date)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []json.RawMessage{}
		for rows.Next() {
			var body []byte
			if err = rows.Scan(&body); err != nil {
				return nil, err
			}
			out = append(out, json.RawMessage(body))
		}
		return out, rows.Err()
	})
}
