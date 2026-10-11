package scheduler

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestWeeklySchedule(t *testing.T) {
	raw := os.Getenv("WTS_SCHEDULER_TEST_URL")
	if raw == "" {
		t.Skip("requires isolated database harness")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() != "127.0.0.1" || u.Path != "/wts_scheduler_test" {
		t.Fatal("requires isolated test database")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, raw)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	admin, err := pgxpool.New(ctx, os.Getenv("WTS_SCHEDULER_TEST_ADMIN_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	// Other tests run sequentially in this isolated database; each fixture defines its own templates.
	if _, err = admin.Exec(ctx, `DELETE FROM scheduler.weekly_templates`); err != nil {
		t.Fatal(err)
	}
	for wid, role := range map[string]string{"weekly-manager": "chief", "weekly-leader": "group-leader", "weekly-leader2": "group-leader", "weekly-a": "formal-member", "weekly-b": "formal-member", "weekly-c": "informal-member"} {
		if _, err = admin.Exec(ctx, `INSERT INTO data.students(sid,name) VALUES($1,$1)`, wid); err != nil {
			t.Fatal(err)
		}
		if _, err = admin.Exec(ctx, `INSERT INTO wts.users(sid,wx) VALUES($1,$1)`, wid); err != nil {
			t.Fatal(err)
		}
		if _, err = admin.Exec(ctx, `INSERT INTO wts.operators(wid,sid,access,female) VALUES($1,$1,$2::wts.access,false)`, wid, role); err != nil {
			t.Fatal(err)
		}
	}
	s := Service{Pool: pool}
	first := time.Now().In(china).AddDate(0, 0, 100)
	for first.Weekday() != time.Friday {
		first = first.AddDate(0, 0, 1)
	}
	date := func(offset int) string { return first.AddDate(0, 0, offset).Format(time.DateOnly) }
	change := func(actor, action string, r Request) Day {
		t.Helper()
		v, e := s.Change(ctx, actor, action, r)
		if e != nil {
			t.Fatal(action, e)
		}
		return v.(Day)
	}
	get := func(value string) Day {
		t.Helper()
		v, e := s.Run(ctx, "weekly-manager", func(w *work) (any, error) { return w.day(value, false) })
		if e != nil {
			t.Fatal(e)
		}
		return v.(Day)
	}
	publish := func(d Day, leader, member string) Day {
		d = change(leader, "assignments", Request{Date: d.Date, Revision: d.Revision, Reason: "weekly test", Assignments: []Assignment{{Wid: leader, Zone: "FX"}, {Wid: member, Zone: "BM", AreaLeader: true}}})
		return change(leader, "publish", Request{Date: d.Date, Revision: d.Revision, Reason: "availability confirmed", OverrideUnavailable: true})
	}
	a := change("weekly-manager", "group", Request{Date: date(0), LeaderWid: "weekly-leader", Wids: []string{"weekly-a"}})
	a = publish(a, "weekly-leader", "weekly-a")
	b := change("weekly-manager", "group", Request{Date: date(1), LeaderWid: "weekly-leader2", Wids: []string{"weekly-b"}})
	b = publish(b, "weekly-leader2", "weekly-b")
	next := get(date(7))
	later := get(date(14))
	if !next.Published || next.Leader.Wid != "weekly-leader" || len(next.Members) != 2 || !memberExists(next, "weekly-a") || !memberExists(later, "weekly-a") {
		t.Fatal("weekday did not repeat")
	}
	for _, m := range next.Members {
		if m.Wid == "weekly-a" && (m.Zone != "BM" || !m.AreaLeader) {
			t.Fatal("area arrangement did not repeat")
		}
	}
	view, e := s.Overview(ctx, "weekly-a", date(14)[:7])
	if e != nil {
		t.Fatal(e)
	}
	visible := false
	for _, d := range view.(Overview).Days {
		if d.Date == date(14) && d.Published {
			visible = true
		}
	}
	if !visible {
		t.Fatal("future duty missing from calendar")
	}
	next = change("weekly-leader", "extra", Request{Date: next.Date, Revision: next.Revision, Wid: "weekly-c", Zone: "FX", Reason: "one-time extra"})
	if memberExists(get(date(14)), "weekly-c") {
		t.Fatal("extra leaked into weekly template")
	}
	next = change("weekly-leader", "remove_extra", Request{Date: next.Date, Revision: next.Revision, Wid: "weekly-c", Reason: "test removal"})
	_ = get(date(8))
	swap, e := s.CreateSwap(ctx, "weekly-a", SwapRequest{DateA: date(7), DateB: date(8)})
	if e != nil {
		t.Fatal(e)
	}
	id := swap.(Swap).ID
	for _, step := range []struct{ actor, action string }{{"weekly-b", "claim"}, {"weekly-leader", "approve"}, {"weekly-leader2", "approve"}} {
		if _, e = s.ReviewSwap(ctx, step.actor, SwapRequest{ID: id, Action: step.action}); e != nil {
			t.Fatal(e)
		}
	}
	if !memberExists(get(date(7)), "weekly-b") || !memberExists(get(date(14)), "weekly-a") {
		t.Fatal("swap changed recurring schedule")
	}
	// Editing and republishing an exchanged date must not turn it into a new template.
	_ = publish(get(date(7)), "weekly-leader", "weekly-b")
	if !memberExists(get(date(14)), "weekly-a") || memberExists(get(date(14)), "weekly-b") {
		t.Fatal("republishing a swap leaked temporary members into recurring schedule")
	}
	var swappedTemplate bool
	if err = admin.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM scheduler.weekly_templates WHERE effective_date=$1::date)`, date(7)).Scan(&swappedTemplate); err != nil || swappedTemplate {
		t.Fatal("dated exception became a template", err)
	}
	a = change("weekly-manager", "group", Request{Date: a.Date, Revision: a.Revision, LeaderWid: "weekly-leader", Wids: []string{"weekly-c"}, Reason: "update Friday roster"})
	a = publish(a, "weekly-leader", "weekly-c")
	if !memberExists(get(date(14)), "weekly-c") || memberExists(get(date(14)), "weekly-a") {
		t.Fatal("template change did not update future Friday")
	}
	if !memberExists(get(date(7)), "weekly-b") {
		t.Fatal("template change overwrote one-time swap")
	}
	d := get(date(21))
	d = change("weekly-manager", "group", Request{Date: d.Date, Revision: d.Revision, LeaderWid: "weekly-leader", Wids: []string{"weekly-b"}, Reason: "new effective version"})
	_ = publish(d, "weekly-leader", "weekly-b")
	a = change("weekly-manager", "group", Request{Date: a.Date, Revision: a.Revision, LeaderWid: "weekly-leader", Wids: []string{"weekly-a"}, Reason: "older version edit"})
	_ = publish(a, "weekly-leader", "weekly-a")
	if !memberExists(get(date(28)), "weekly-b") {
		t.Fatal("older edit overwrote newer effective version")
	}
	if _, e = s.Change(ctx, "weekly-leader2", "assignments", Request{Date: date(28), Revision: get(date(28)).Revision}); e == nil {
		t.Fatal("unrelated leader edited weekday roster")
	}
	if _, e = s.Run(ctx, "weekly-manager", func(w *work) (any, error) { return w.day(date(56), false) }); e == nil {
		t.Fatal("weekly roster extended beyond one calendar month")
	}
	// Saving creates all occurrences, before anyone opens the month/calendar.
	var materialized bool
	if err = admin.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM scheduler.duty_days WHERE duty_date=$1::date)`, date(35)).Scan(&materialized); err != nil || !materialized {
		t.Fatal("save did not materialize the month", err)
	}
	// Older lazy templates can backfill a missing historical occurrence, but never overwrite it.
	history := time.Now().In(china).AddDate(0, 0, -21).Format(time.DateOnly)
	historical := change("weekly-manager", "group", Request{Date: history, LeaderWid: "weekly-leader", Wids: []string{"weekly-a"}})
	_ = publish(historical, "weekly-leader", "weekly-a")
	historyNext := time.Now().In(china).AddDate(0, 0, -14).Format(time.DateOnly)
	if _, err = admin.Exec(ctx, `DELETE FROM scheduler.duty_days WHERE duty_date=$1::date`, historyNext); err != nil {
		t.Fatal(err)
	}
	oldDay := get(historyNext)
	if !oldDay.Published || !memberExists(oldDay, "weekly-a") {
		t.Fatal("unvisited historical occurrence missing")
	}
	historical = get(history)
	_ = change("weekly-manager", "group", Request{Date: history, Revision: historical.Revision, LeaderWid: "weekly-leader", Wids: []string{"weekly-c"}, Reason: "historical edit"})
	if !memberExists(get(historyNext), "weekly-a") {
		t.Fatal("existing history changed with a later template edit")
	}
}
