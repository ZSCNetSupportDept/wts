package scheduler

import (
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"time"
)

func (w *work) saveWeekly(d Day) error {
	var exception bool
	if err := w.tx.QueryRow(w.ctx, `SELECT weekly_exception FROM scheduler.duty_days WHERE duty_date=$1::date`, d.Date).Scan(&exception); err != nil {
		return err
	}
	// A dated swap/extra remains independent even when it is edited or republished.
	if exception {
		return nil
	}
	members := []Assignment{}
	for _, m := range d.Members {
		if !m.Extra {
			members = append(members, Assignment{Wid: m.Wid, Zone: m.Zone, AreaLeader: m.AreaLeader})
		}
	}
	body, err := json.Marshal(members)
	if err != nil {
		return err
	}
	_, err = w.tx.Exec(w.ctx, `INSERT INTO scheduler.weekly_templates(effective_date,weekday,leader_wid,published,revision,members,updated_by,ends_on)
 VALUES($1::date,extract(isodow FROM $1::date)::int,$2,$3,$4,$5::jsonb,$6,($1::date+INTERVAL '1 month')::date)
 ON CONFLICT(effective_date) DO UPDATE SET leader_wid=EXCLUDED.leader_wid,published=EXCLUDED.published,
 revision=EXCLUDED.revision,members=EXCLUDED.members,updated_by=EXCLUDED.updated_by,updated_at=NOW()`, d.Date, d.Leader.Wid, d.Published, d.Revision, string(body), w.actor.Wid)
	if err != nil {
		return err
	}
	_, err = w.tx.Exec(w.ctx, `UPDATE scheduler.duty_days SET weekly_source=NULL,weekly_revision=0,weekly_exception=false WHERE duty_date=$1::date`, d.Date)
	if err != nil {
		return err
	}
	var end string
	if err = w.tx.QueryRow(w.ctx, `SELECT to_char(ends_on,'YYYY-MM-DD') FROM scheduler.weekly_templates WHERE effective_date=$1::date`, d.Date).Scan(&end); err != nil {
		return err
	}
	start, _ := parseDate(d.Date)
	for next := start.AddDate(0, 0, 7); next.Format(time.DateOnly) < end; next = next.AddDate(0, 0, 7) {
		date := next.Format(time.DateOnly)
		delete(w.weeklyChecked, date)
		if err = w.ensureWeekly(date); err != nil {
			return err
		}
	}
	return nil
}

// Latest effective weekday template supplies future dates. One-off exceptions stay frozen.
func (w *work) ensureWeekly(date string) error {
	if w.weeklyChecked == nil {
		w.weeklyChecked = map[string]bool{}
	}
	if w.weeklyChecked[date] {
		return nil
	}
	w.weeklyChecked[date] = true
	var source, leader, actor string
	var published bool
	var revision int
	var body []byte
	err := w.tx.QueryRow(w.ctx, `SELECT to_char(effective_date,'YYYY-MM-DD'),leader_wid,published,revision,members,updated_by
 FROM scheduler.weekly_templates WHERE effective_date<=$1::date AND ends_on>$1::date AND weekday=extract(isodow FROM $1::date)::int
 ORDER BY effective_date DESC LIMIT 1`, date).Scan(&source, &leader, &published, &revision, &body, &actor)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if source == date {
		return nil
	}
	var currentSource *string
	var currentRevision int
	var exception bool
	err = w.tx.QueryRow(w.ctx, `SELECT to_char(weekly_source,'YYYY-MM-DD'),weekly_revision,weekly_exception FROM scheduler.duty_days WHERE duty_date=$1::date FOR UPDATE`, date).Scan(&currentSource, &currentRevision, &exception)
	existing := err == nil
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if existing && (date < time.Now().In(china).Format(time.DateOnly) || currentSource == nil || exception || (*currentSource == source && currentRevision == revision)) {
		return nil
	}
	members := []Assignment{}
	if err = json.Unmarshal(body, &members); err != nil {
		return err
	}
	if _, err = w.tx.Exec(w.ctx, `INSERT INTO scheduler.duty_days(duty_date,leader_wid,published,updated_by,weekly_source,weekly_revision)
 VALUES($1::date,$2,$3,$4,$5::date,$6) ON CONFLICT(duty_date) DO UPDATE SET leader_wid=EXCLUDED.leader_wid,
 published=EXCLUDED.published,updated_by=EXCLUDED.updated_by,weekly_source=EXCLUDED.weekly_source,
 weekly_revision=EXCLUDED.weekly_revision,revision=scheduler.duty_days.revision+1,updated_at=NOW()`, date, leader, published, actor, source, revision); err != nil {
		return err
	}
	if _, err = w.tx.Exec(w.ctx, `DELETE FROM scheduler.duty_members WHERE duty_date=$1::date`, date); err != nil {
		return err
	}
	for _, m := range members {
		if _, err = w.tx.Exec(w.ctx, `INSERT INTO scheduler.duty_members(duty_date,wid,zone,area_leader) VALUES($1::date,$2,$3,$4)`, date, m.Wid, m.Zone, m.AreaLeader); err != nil {
			return err
		}
	}
	_, err = w.tx.Exec(w.ctx, `UPDATE scheduler.swap_requests SET status='stale' WHERE status IN ('open','pending') AND (date_a=$1::date OR date_b=$1::date)`, date)
	return err
}
