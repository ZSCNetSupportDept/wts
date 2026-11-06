package scheduler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct{ Pool *pgxpool.Pool }
type work struct {
	ctx           context.Context
	tx            pgx.Tx
	actor         Person
	weeklyChecked map[string]bool
}

// Resolve the current database identity on every request. Debug JWT bypass is not used.
func (s Service) Run(ctx context.Context, openID string, fn func(*work) (any, error)) (any, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())
	// Serialize weekday templates and occurrence snapshots before taking identity locks.
	if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(83629,1)`); err != nil {
		return nil, err
	}
	w := &work{ctx: ctx, tx: tx}
	var active bool
	err = tx.QueryRow(ctx, `SELECT o.wid, COALESCE(s.name,''), o.access::text,u.op
        FROM wts.users u JOIN wts.operators o ON o.sid=u.sid
        LEFT JOIN data.students s ON s.sid=u.sid
        WHERE u.wx=$1 FOR SHARE OF u,o`, openID).
		Scan(&w.actor.Wid, &w.actor.Name, &w.actor.Access, &active)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fail(403, "当前登录身份未关联在职成员，旧登录可能已失效，请重新使用微信登录")
	}
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, fail(403, "该成员的在职权限已停用，请联系管理员")
	}
	if !isOperator(w.actor.Access) || w.actor.Wid == "-1" || w.actor.Wid == "-2" {
		return nil, fail(403, "无排班权限")
	}
	if _, err = tx.Exec(ctx, `SELECT set_config('wts.wx',$1,true)`, openID); err != nil {
		return nil, err
	}
	result, err := fn(w)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return result, nil
}

func (w *work) person(wid string) (Person, error) {
	var p Person
	err := w.tx.QueryRow(w.ctx, `SELECT o.wid,COALESCE(s.name,''),o.access::text
        FROM wts.operators o JOIN wts.users u ON u.sid=o.sid
        LEFT JOIN data.students s ON s.sid=o.sid WHERE o.wid=$1 AND u.op FOR SHARE OF o,u`, wid).
		Scan(&p.Wid, &p.Name, &p.Access)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, fail(400, "工号不存在或成员已离职")
	}
	if err == nil && (!isOperator(p.Access) || wid == "-1" || wid == "-2") {
		return p, fail(400, "该成员不能参与排班")
	}
	return p, err
}

func (w *work) day(date string, lock bool) (Day, error) {
	if err := w.ensureWeekly(date); err != nil {
		return Day{}, err
	}
	var d Day
	d.Members = []Member{}
	query := `SELECT to_char(d.duty_date,'YYYY-MM-DD'),o.wid,COALESCE(s.name,''),o.access::text,d.published,d.revision
        FROM scheduler.duty_days d JOIN wts.operators o ON o.wid=d.leader_wid
        LEFT JOIN data.students s ON s.sid=o.sid WHERE d.duty_date=$1::date`
	if lock {
		query += ` FOR UPDATE OF d`
	}
	err := w.tx.QueryRow(w.ctx, query, date).Scan(&d.Date, &d.Leader.Wid, &d.Leader.Name, &d.Leader.Access, &d.Published, &d.Revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, fail(404, "当天尚未配置小组")
	}
	if err != nil {
		return d, err
	}
	rows, err := w.tx.Query(w.ctx, `SELECT o.wid,COALESCE(s.name,''),o.access::text,m.zone,m.area_leader,m.extra,
        COALESCE(e.available,EXISTS(SELECT 1 FROM scheduler.freeday f WHERE f.wid=o.wid
          AND f.free_at::text=extract(isodow FROM m.duty_date)::int::text))
        FROM scheduler.duty_members m JOIN wts.operators o ON o.wid=m.wid
        LEFT JOIN data.students s ON s.sid=o.sid
        LEFT JOIN scheduler.availability_exceptions e ON e.wid=m.wid AND e.duty_date=m.duty_date
        WHERE m.duty_date=$1::date ORDER BY m.extra,m.zone,m.area_leader DESC,o.wid`, date)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Member
		if err = rows.Scan(&m.Wid, &m.Name, &m.Access, &m.Zone, &m.AreaLeader, &m.Extra, &m.Available); err != nil {
			return d, err
		}
		d.Members = append(d.Members, m)
	}
	return d, rows.Err()
}

func (s Service) Overview(ctx context.Context, openID, month string) (any, error) {
	start, err := time.ParseInLocation("2006-01", month, china)
	if err != nil || start.Format("2006-01") != month {
		return nil, fail(400, "月份必须为 YYYY-MM")
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		out := Overview{Me: w.actor, Manager: isManager(w.actor.Access), Days: []Day{}, FreeDays: []int{}, Exceptions: []Exception{}}
		for d := start; d.Before(start.AddDate(0, 1, 0)); d = d.AddDate(0, 0, 1) {
			if err := w.ensureWeekly(d.Format(time.DateOnly)); err != nil {
				return nil, err
			}
		}
		rows, err := w.tx.Query(ctx, `SELECT to_char(duty_date,'YYYY-MM-DD') FROM scheduler.duty_days
            WHERE duty_date >= $1::date AND duty_date < $2::date
              AND (published OR $3 OR ($4 AND leader_wid=$5)) ORDER BY duty_date`,
			start.Format(time.DateOnly), start.AddDate(0, 1, 0).Format(time.DateOnly), out.Manager, w.actor.Access == "group-leader", w.actor.Wid)
		if err != nil {
			return nil, err
		}
		dates := []string{}
		for rows.Next() {
			var date string
			if err = rows.Scan(&date); err != nil {
				rows.Close()
				return nil, err
			}
			dates = append(dates, date)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		for _, date := range dates {
			d, e := w.day(date, false)
			if e != nil {
				return nil, e
			}
			out.Days = append(out.Days, d)
		}
		rows, err = w.tx.Query(ctx, `SELECT free_at::text::int FROM scheduler.freeday WHERE wid=$1 ORDER BY free_at`, w.actor.Wid)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var day int
			if err = rows.Scan(&day); err != nil {
				rows.Close()
				return nil, err
			}
			out.FreeDays = append(out.FreeDays, day)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return nil, err
		}
		rows, err = w.tx.Query(ctx, `SELECT to_char(duty_date,'YYYY-MM-DD'),available FROM scheduler.availability_exceptions
            WHERE wid=$1 AND duty_date >= $2::date AND duty_date < $3::date ORDER BY duty_date`, w.actor.Wid, start.Format(time.DateOnly), start.AddDate(0, 1, 0).Format(time.DateOnly))
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var e Exception
			if err = rows.Scan(&e.Date, &e.Available); err != nil {
				return nil, err
			}
			out.Exceptions = append(out.Exceptions, e)
		}
		return out, rows.Err()
	})
}

func (s Service) Staff(ctx context.Context, openID, date string) (any, error) {
	if _, err := parseDate(date); err != nil {
		return nil, err
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		if !isManager(w.actor.Access) {
			d, err := w.day(date, false)
			if err != nil {
				return nil, err
			}
			if !canArrange(w.actor, d) {
				return nil, fail(403, "只能管理自己负责日期的排班")
			}
		}
		rows, err := w.tx.Query(ctx, `SELECT o.wid,COALESCE(s.name,''),o.access::text,
            COALESCE(e.available,EXISTS(SELECT 1 FROM scheduler.freeday f WHERE f.wid=o.wid AND f.free_at::text=extract(isodow FROM $1::date)::int::text))
            FROM wts.operators o JOIN wts.users u ON u.sid=o.sid LEFT JOIN data.students s ON s.sid=o.sid
            LEFT JOIN scheduler.availability_exceptions e ON e.wid=o.wid AND e.duty_date=$1::date
            WHERE u.op AND o.access IN ('chief','dev','api','group-leader','formal-member','informal-member')
              AND o.wid NOT IN ('-1','-2') ORDER BY o.wid`, date)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		out := []Member{}
		for rows.Next() {
			var p Member
			if err = rows.Scan(&p.Wid, &p.Name, &p.Access, &p.Available); err != nil {
				return nil, err
			}
			out = append(out, p)
		}
		return out, rows.Err()
	})
}

type Request struct {
	Date                string       `json:"date"`
	Revision            int          `json:"revision"`
	LeaderWid           string       `json:"leader_wid"`
	Wids                []string     `json:"wids"`
	Assignments         []Assignment `json:"assignments"`
	Wid                 string       `json:"wid"`
	Zone                string       `json:"zone"`
	Reason              string       `json:"reason"`
	OverrideUnavailable bool         `json:"override_unavailable"`
	Mode                string       `json:"mode"`
}

// Lock the date even before its first row exists; concurrent creators cannot overwrite it.
func (w *work) lockDate(date string) error {
	_, err := w.tx.Exec(w.ctx, `SELECT pg_advisory_xact_lock(83627,($1::date-DATE '2000-01-01')::int)`, date)
	return err
}
func (w *work) record(before Day, action, reason string) (Day, error) {
	if action == "extra" || action == "remove_extra" || action == "swap_complete" {
		if _, err := w.tx.Exec(w.ctx, `UPDATE scheduler.duty_days SET weekly_exception=true WHERE duty_date=$1::date`, before.Date); err != nil {
			return Day{}, err
		}
	}
	_, err := w.tx.Exec(w.ctx, `UPDATE scheduler.duty_days SET revision=revision+1,updated_by=$2,updated_at=now() WHERE duty_date=$1::date`, before.Date, w.actor.Wid)
	if err != nil {
		return Day{}, err
	}
	// Pending approvals refer to exact snapshots and must be requested again after any change.
	_, err = w.tx.Exec(w.ctx, `UPDATE scheduler.swap_requests SET status='stale' WHERE status IN ('open','pending') AND (date_a=$1::date OR date_b=$1::date)`, before.Date)
	if err != nil {
		return Day{}, err
	}
	after, err := w.day(before.Date, false)
	if err != nil {
		return after, err
	}
	oldJSON, err := json.Marshal(before)
	if err != nil {
		return after, err
	}
	newJSON, err := json.Marshal(after)
	if err != nil {
		return after, err
	}
	_, err = w.tx.Exec(w.ctx, `INSERT INTO scheduler.audit_log(duty_date,actor_wid,action,reason,before_state,after_state) VALUES($1::date,$2,$3,$4,$5::jsonb,$6::jsonb)`, before.Date, w.actor.Wid, action, reason, string(oldJSON), string(newJSON))
	if err == nil && (action == "group" || action == "assignments" || action == "publish") {
		err = w.saveWeekly(after)
	}
	return after, err
}

func (s Service) Change(ctx context.Context, openID, action string, r Request) (any, error) {
	if _, err := parseDate(r.Date); err != nil {
		return nil, err
	}
	r.Reason = strings.TrimSpace(r.Reason)
	if len(r.Reason) > 1000 {
		return nil, fail(400, "原因过长")
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		if err := w.lockDate(r.Date); err != nil {
			return nil, err
		}
		d, err := w.day(r.Date, true)
		if action == "group" {
			if !isManager(w.actor.Access) {
				return nil, fail(403, "只有科长/dev/api能配置每日小组")
			}
			if err != nil {
				var e *Error
				if !errors.As(err, &e) || e.Code != 404 {
					return nil, err
				}
				if r.Revision != 0 {
					return nil, fail(409, "当天小组不存在，请刷新")
				}
				d = Day{Date: r.Date, Members: []Member{}}
			} else {
				if err = checkRevision(d, r.Revision); err != nil {
					return nil, err
				}
			}
			if d.Published && r.Reason == "" {
				return nil, fail(400, "修改已发布的小组需要填写原因")
			}
			leader, err := w.person(r.LeaderWid)
			if err != nil {
				return nil, err
			}
			if leader.Access != "group-leader" {
				return nil, fail(400, "请选择当前具有组长权限的值班组长")
			}
			if len(r.Wids) > 200 {
				return nil, fail(400, "小组人数过多")
			}
			seen := map[string]bool{leader.Wid: true}
			members := []string{leader.Wid}
			for _, wid := range r.Wids {
				if seen[wid] {
					continue
				}
				if _, err = w.person(wid); err != nil {
					return nil, err
				}
				seen[wid] = true
				members = append(members, wid)
			}
			_, err = w.tx.Exec(ctx, `INSERT INTO scheduler.duty_days(duty_date,leader_wid,revision,updated_by) VALUES($1::date,$2,1,$3)
                ON CONFLICT(duty_date) DO UPDATE SET leader_wid=EXCLUDED.leader_wid,published=false`, r.Date, leader.Wid, w.actor.Wid)
			if err != nil {
				return nil, err
			}
			_, err = w.tx.Exec(ctx, `DELETE FROM scheduler.duty_members WHERE duty_date=$1::date AND NOT extra AND NOT(wid=ANY($2::text[]))`, r.Date, members)
			if err != nil {
				return nil, err
			}
			for _, wid := range members {
				if _, err = w.tx.Exec(ctx, `INSERT INTO scheduler.duty_members(duty_date,wid) VALUES($1::date,$2) ON CONFLICT(duty_date,wid) DO UPDATE SET extra=false`, r.Date, wid); err != nil {
					return nil, err
				}
			}
			return w.record(d, action, r.Reason)
		}
		if err != nil {
			return nil, err
		}
		if !canArrange(w.actor, d) {
			return nil, fail(403, "组长只能操作自己负责的日期")
		}
		if err = checkRevision(d, r.Revision); err != nil {
			return nil, err
		}
		if action == "suggest" {
			return w.suggest(d, r.Mode)
		}
		if d.Published && r.Reason == "" {
			return nil, fail(400, "修改已发布排班需要填写原因")
		}
		switch action {
		case "assignments":
			if err = validateAssignments(d, r.Assignments); err != nil {
				return nil, err
			}
			for _, a := range r.Assignments {
				if _, err = w.person(a.Wid); err != nil {
					return nil, err
				}
				_, err = w.tx.Exec(ctx, `UPDATE scheduler.duty_members SET zone=$3,area_leader=$4 WHERE duty_date=$1::date AND wid=$2 AND NOT extra`, r.Date, a.Wid, a.Zone, a.AreaLeader)
				if err != nil {
					return nil, err
				}
			}
			if _, err = w.tx.Exec(ctx, `UPDATE scheduler.duty_days SET published=false WHERE duty_date=$1::date`, r.Date); err != nil {
				return nil, err
			}
		case "publish":
			for _, m := range d.Members {
				if _, err = w.person(m.Wid); err != nil {
					return nil, err
				}
				if !validZone(m.Zone) {
					return nil, fail(400, "发布前必须为所有成员选择片区")
				}
				if !m.Available && (!r.OverrideUnavailable || r.Reason == "") {
					return nil, fail(400, "存在未填报空闲或当天不可用的成员；特殊安排需勾选确认并填写原因")
				}
			}
			if p, e := w.person(d.Leader.Wid); e != nil {
				return nil, e
			} else if p.Access != "group-leader" {
				return nil, fail(400, "值班组长权限已变更，请重新配置小组")
			}
			if _, err = w.tx.Exec(ctx, `UPDATE scheduler.duty_days SET published=true WHERE duty_date=$1::date`, r.Date); err != nil {
				return nil, err
			}
		case "extra":
			if !validZone(r.Zone) {
				return nil, fail(400, "请选择有效片区")
			}
			if _, err = w.person(r.Wid); err != nil {
				return nil, err
			}
			for _, m := range d.Members {
				if m.Wid == r.Wid {
					return nil, fail(409, "该成员已在本班次中")
				}
			}
			_, err = w.tx.Exec(ctx, `INSERT INTO scheduler.duty_members(duty_date,wid,zone,extra) VALUES($1::date,$2,$3,true)`, r.Date, r.Wid, r.Zone)
			if err != nil {
				return nil, err
			}
		case "remove_extra":
			tag, e := w.tx.Exec(ctx, `DELETE FROM scheduler.duty_members WHERE duty_date=$1::date AND wid=$2 AND extra`, r.Date, r.Wid)
			if e != nil {
				return nil, e
			}
			if tag.RowsAffected() != 1 {
				return nil, fail(400, "只能移除蹭班成员")
			}
		default:
			return nil, fail(400, "未知操作")
		}
		result, err := w.record(d, action, r.Reason)
		if err != nil {
			return nil, err
		}
		leaders := map[string]bool{}
		for _, m := range result.Members {
			if m.AreaLeader {
				if leaders[m.Zone] {
					return nil, fail(400, "每个片区只能有一位负责人")
				}
				leaders[m.Zone] = true
			}
		}
		return result, nil
	})
}

type AvailabilityRequest struct {
	FreeDays  []int  `json:"free_days"`
	Date      string `json:"date"`
	Available *bool  `json:"available"`
}

func (s Service) Availability(ctx context.Context, openID string, r AvailabilityRequest) (any, error) {
	seen := map[int]bool{}
	for _, d := range r.FreeDays {
		if d < 1 || d > 7 || seen[d] {
			return nil, fail(400, "空闲星期必须为不重复的 1 至 7")
		}
		seen[d] = true
	}
	if r.Date != "" {
		if _, err := parseDate(r.Date); err != nil {
			return nil, err
		}
	}
	return s.Run(ctx, openID, func(w *work) (any, error) {
		// The actor row is locked above. A second update for this person waits, avoiding lost updates.
		if _, err := w.tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,83628))`, w.actor.Wid); err != nil {
			return nil, err
		}
		if _, err := w.tx.Exec(ctx, `DELETE FROM scheduler.freeday WHERE wid=$1`, w.actor.Wid); err != nil {
			return nil, err
		}
		for _, d := range r.FreeDays {
			if _, err := w.tx.Exec(ctx, `INSERT INTO scheduler.freeday(wid,free_at) VALUES($1,$2::scheduler.weekday)`, w.actor.Wid, string(rune('0'+d))); err != nil {
				return nil, err
			}
		}
		if r.Date != "" {
			if r.Available == nil {
				if _, err := w.tx.Exec(ctx, `DELETE FROM scheduler.availability_exceptions WHERE wid=$1 AND duty_date=$2::date`, w.actor.Wid, r.Date); err != nil {
					return nil, err
				}
			} else {
				if _, err := w.tx.Exec(ctx, `INSERT INTO scheduler.availability_exceptions(wid,duty_date,available) VALUES($1,$2::date,$3)
                    ON CONFLICT(wid,duty_date) DO UPDATE SET available=EXCLUDED.available`, w.actor.Wid, r.Date, *r.Available); err != nil {
					return nil, err
				}
			}
		}
		body, err := json.Marshal(r)
		if err != nil {
			return nil, err
		}
		_, err = w.tx.Exec(ctx, `INSERT INTO scheduler.audit_log(actor_wid,action,after_state) VALUES($1,'availability',$2::jsonb)`, w.actor.Wid, string(body))
		return map[string]bool{"saved": true}, err
	})
}

type Suggestion struct {
	PreviousDate string         `json:"previous_date"`
	Assignments  []Assignment   `json:"assignments"`
	Quotas       map[string]int `json:"quotas"`
}

func (w *work) suggest(d Day, mode string) (any, error) {
	var date string
	err := w.tx.QueryRow(w.ctx, `SELECT to_char(duty_date,'YYYY-MM-DD') FROM scheduler.duty_days
        WHERE leader_wid=$1 AND published AND duty_date<$2::date ORDER BY duty_date DESC LIMIT 1`, d.Leader.Wid, d.Date).Scan(&date)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fail(400, "该组长尚无上一次已发布排班，请先手动安排")
	}
	if err != nil {
		return nil, err
	}
	previous, err := w.day(date, false)
	if err != nil {
		return nil, err
	}
	out := Suggestion{PreviousDate: date, Assignments: []Assignment{}, Quotas: map[string]int{}}
	old := map[string]Member{}
	for _, m := range previous.Members {
		if !m.Extra {
			old[m.Wid] = m
			out.Quotas[m.Zone]++
		}
	}
	for _, m := range d.Members {
		if !m.Extra {
			a := Assignment{Wid: m.Wid, Zone: m.Zone, AreaLeader: m.AreaLeader}
			if p, ok := old[m.Wid]; ok {
				a.Zone = p.Zone
				a.AreaLeader = p.AreaLeader
			}
			out.Assignments = append(out.Assignments, a)
		}
	}
	if mode == "reuse" {
		return out, nil
	}
	if mode != "random" {
		return nil, fail(400, "请选择沿用或随机排班")
	}
	// Keep all non-interns fixed; only shuffle the intern pool against previous total zone capacities.
	people := map[string]Member{}
	for _, m := range d.Members {
		people[m.Wid] = m
	}
	for i, a := range out.Assignments {
		m := people[a.Wid]
		if m.Access != "informal-member" {
			if validZone(m.Zone) {
				out.Assignments[i].Zone = m.Zone
				out.Assignments[i].AreaLeader = m.AreaLeader
			}
		}
	}
	assignments, err := randomize(out.Assignments, people, out.Quotas)
	if err != nil {
		return nil, err
	}
	out.Assignments = assignments
	return out, nil
}
func randomize(assignments []Assignment, people map[string]Member, quotas map[string]int) ([]Assignment, error) {
	remaining := map[string]int{}
	for zone, count := range quotas {
		remaining[zone] = count
	}
	indexes := []int{}
	for i, a := range assignments {
		if people[a.Wid].Access == "informal-member" {
			indexes = append(indexes, i)
		} else {
			if !validZone(a.Zone) {
				return nil, fail(400, "请先为正式成员、组长等固定人员选择片区并保存")
			}
			remaining[a.Zone]--
			if remaining[a.Zone] < 0 {
				return nil, fail(400, "固定人员已超过上次对应片区人数，请手动调整")
			}
		}
	}
	slots := []string{}
	for _, zone := range []string{"FX", "BM", "DM", "QT", "XHAB", "XHCD", "ZH", "other"} {
		for n := 0; n < remaining[zone]; n++ {
			slots = append(slots, zone)
		}
	}
	if len(slots) != len(indexes) {
		return nil, fail(400, "当前小组人数与上次片区人数不匹配，无法随机排班")
	}
	for i := len(slots) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		j := int(n.Int64())
		slots[i], slots[j] = slots[j], slots[i]
	}
	out := append([]Assignment(nil), assignments...)
	for j, i := range indexes {
		out[i].Zone = slots[j]
		out[i].AreaLeader = false
	}
	return out, nil
}
