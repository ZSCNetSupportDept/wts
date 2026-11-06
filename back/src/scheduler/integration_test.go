package scheduler

import (
	"context"
	"errors"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSchedulerDatabase(t *testing.T) {
	connection := os.Getenv("WTS_SCHEDULER_TEST_URL")
	if connection == "" {
		t.Skip("run back/scripts/test-scheduler.mjs for isolated PostgreSQL integration tests")
	}
	parsed, err := url.Parse(connection)
	if err != nil || parsed.Hostname() != "127.0.0.1" || parsed.Path != "/wts_scheduler_test" {
		t.Fatal("integration tests require the isolated local wts_scheduler_test database")
	}
	adminURL := os.Getenv("WTS_SCHEDULER_TEST_ADMIN_URL")
	adminParsed, e := url.Parse(adminURL)
	if e != nil || adminParsed.Host != parsed.Host || adminParsed.Path != parsed.Path {
		t.Fatal("admin connection must point to the same isolated test database")
	}
	ctx := context.Background()
	admin, err := pgxpool.New(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	service := Service{Pool: pool}
	roles := map[string]string{"101": "group-leader", "102": "group-leader", "103": "group-leader", "201": "formal-member", "202": "formal-member", "301": "informal-member", "302": "informal-member", "900": "chief", "901": "api", "902": "dev", "903": "pre-member"}
	for wid, role := range roles {
		_, err = admin.Exec(ctx, `INSERT INTO data.students(sid,name) VALUES($1,$1);`, wid)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = admin.Exec(ctx, `INSERT INTO wts.users(sid,wx) VALUES($1,$1)`, wid); err != nil {
			t.Fatal(err)
		}
		if _, err = admin.Exec(ctx, `INSERT INTO wts.operators(wid,sid,access,female) VALUES($1,$1,$2::wts.access,false)`, wid, role); err != nil {
			t.Fatal(err)
		}
	}
	dayA := time.Now().In(china).AddDate(0, 0, 4).Format(time.DateOnly)
	dayB := time.Now().In(china).AddDate(0, 0, 5).Format(time.DateOnly)
	dayC := time.Now().In(china).AddDate(0, 0, 6).Format(time.DateOnly)
	previous := time.Now().In(china).AddDate(0, 0, 1).Format(time.DateOnly)
	code := func(err error, want int) {
		t.Helper()
		var e *Error
		if !errors.As(err, &e) || e.Code != want {
			t.Fatalf("wanted %d, got %v", want, err)
		}
	}
	change := func(actor, action string, r Request) Day {
		t.Helper()
		value, err := service.Change(ctx, actor, action, r)
		if err != nil {
			t.Fatalf("%s: %v", action, err)
		}
		return value.(Day)
	}
	published := func(actor string, d Day, assignments []Assignment) Day {
		t.Helper()
		d = change(actor, "assignments", Request{Date: d.Date, Revision: d.Revision, Assignments: assignments})
		return change(actor, "publish", Request{Date: d.Date, Revision: d.Revision, OverrideUnavailable: true, Reason: "test confirmed availability"})
	}
	a := change("900", "group", Request{Date: dayA, LeaderWid: "101", Wids: []string{"201", "301", "302"}})
	b := change("901", "group", Request{Date: dayB, LeaderWid: "102", Wids: []string{"202"}})
	c := change("902", "group", Request{Date: dayC, LeaderWid: "103", Wids: []string{"301"}})
	t.Run("date and role boundaries", func(t *testing.T) {
		_, err := service.Change(ctx, "101", "group", Request{Date: dayA, Revision: a.Revision, LeaderWid: "101"})
		code(err, 403)
		_, err = service.Change(ctx, "102", "extra", Request{Date: dayA, Revision: a.Revision, Wid: "202", Zone: "FX"})
		code(err, 403)
		_, err = service.Overview(ctx, "903", dayA[:7])
		code(err, 403)
		_, err = service.Overview(ctx, "nonexistent", dayA[:7])
		code(err, 403)
		_, err = service.Staff(ctx, "201", dayA)
		code(err, 403)
		value, err := service.Overview(ctx, "201", dayA[:7])
		if err != nil {
			t.Fatal(err)
		}
		if len(value.(Overview).Days) != 0 {
			t.Fatal("member can see drafts")
		}
		_, err = service.Change(ctx, "101", "extra", Request{Date: dayA, Revision: a.Revision - 1, Wid: "202", Zone: "FX"})
		code(err, 409)
	})
	a = published("101", a, []Assignment{{Wid: "101", Zone: "FX"}, {Wid: "201", Zone: "BM", AreaLeader: true}, {Wid: "301", Zone: "FX"}, {Wid: "302", Zone: "BM"}})
	b = published("102", b, []Assignment{{Wid: "102", Zone: "FX"}, {Wid: "202", Zone: "BM"}})
	c = published("103", c, []Assignment{{Wid: "103", Zone: "FX"}, {Wid: "301", Zone: "BM"}})
	prior := change("900", "group", Request{Date: previous, LeaderWid: "101", Wids: []string{"201", "301", "302"}})
	prior = published("101", prior, []Assignment{{Wid: "101", Zone: "FX"}, {Wid: "201", Zone: "BM", AreaLeader: true}, {Wid: "301", Zone: "BM"}, {Wid: "302", Zone: "FX"}})
	t.Run("memory and intern randomization", func(t *testing.T) {
		value, err := service.Change(ctx, "101", "suggest", Request{Date: dayA, Revision: a.Revision, Mode: "reuse"})
		if err != nil {
			t.Fatal(err)
		}
		if value.(Suggestion).PreviousDate != prior.Date {
			t.Fatal("wrong previous roster")
		}
		value, err = service.Change(ctx, "101", "suggest", Request{Date: dayA, Revision: a.Revision, Mode: "random"})
		if err != nil {
			t.Fatal(err)
		}
		counts := map[string]int{}
		for _, row := range value.(Suggestion).Assignments {
			counts[row.Zone]++
			if row.Wid == "201" && row.Zone != "BM" {
				t.Fatal("formal member moved")
			}
		}
		if counts["FX"] != 2 || counts["BM"] != 2 {
			t.Fatal(counts)
		}
	})
	t.Run("extra scope and one-time nature", func(t *testing.T) {
		a = change("101", "extra", Request{Date: dayA, Revision: a.Revision, Wid: "202", Zone: "FX", Reason: "extra"})
		_, err := service.Change(ctx, "101", "extra", Request{Date: dayA, Revision: a.Revision, Wid: "202", Zone: "FX", Reason: "extra"})
		code(err, 409)
		a = change("101", "remove_extra", Request{Date: dayA, Revision: a.Revision, Wid: "202", Reason: "remove"})
	})
	t.Run("two leader swap approval", func(t *testing.T) {
		value, err := service.CreateSwap(ctx, "201", SwapRequest{DateA: dayA, DateB: dayB, Reason: "one-time swap"})
		if err != nil {
			t.Fatal(err)
		}
		request := value.(Swap)
		list, listErr := service.Swaps(ctx, "101", dayA[:7])
		if listErr != nil || len(list.([]Swap)) != 1 || list.([]Swap)[0].CanReview {
			t.Fatal("leader must wait for member acceptance", listErr)
		}
		list, listErr = service.Swaps(ctx, "301", dayA[:7])
		if listErr != nil || len(list.([]Swap)) != 0 {
			t.Fatal("unrelated member can read swap requests", listErr)
		}
		list, listErr = service.Swaps(ctx, "202", dayB[:7])
		if listErr != nil || len(list.([]Swap)) != 1 || !list.([]Swap)[0].CanClaim {
			t.Fatal("expected date member cannot view hall", listErr)
		}
		_, err = service.ReviewSwap(ctx, "101", SwapRequest{ID: request.ID, Action: "approve"})
		code(err, 409)
		_, err = service.ReviewSwap(ctx, "301", SwapRequest{ID: request.ID, Action: "claim"})
		code(err, 403)
		value, err = service.ReviewSwap(ctx, "202", SwapRequest{ID: request.ID, Action: "claim"})
		if err != nil {
			t.Fatal(err)
		}
		if value.(Swap).Status != "pending" || value.(Swap).TargetWid != "202" {
			t.Fatal("claim did not enter approval queue")
		}
		_, err = service.ReviewSwap(ctx, "900", SwapRequest{ID: request.ID, Action: "approve"})
		code(err, 403)
		_, err = service.ReviewSwap(ctx, "103", SwapRequest{ID: request.ID, Action: "approve"})
		code(err, 403)
		value, err = service.ReviewSwap(ctx, "101", SwapRequest{ID: request.ID, Action: "approve"})
		if err != nil {
			t.Fatal(err)
		}
		if value.(Swap).Status != "pending" {
			t.Fatal("one approval applied swap")
		}
		var stillPresent bool
		if err = admin.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM scheduler.duty_members WHERE duty_date=$1::date AND wid='201')`, dayA).Scan(&stillPresent); err != nil || !stillPresent {
			t.Fatal("swap executed too early", err)
		}
		value, err = service.ReviewSwap(ctx, "102", SwapRequest{ID: request.ID, Action: "approve"})
		if err != nil {
			t.Fatal(err)
		}
		if value.(Swap).Status != "approved" {
			t.Fatal("swap incomplete")
		}
		var total int
		if err = admin.QueryRow(ctx, `SELECT count(*) FROM scheduler.duty_members WHERE (duty_date=$1::date AND wid='202' AND zone='BM') OR (duty_date=$2::date AND wid='201' AND zone='BM') OR (duty_date=$3::date AND wid='201')`, dayA, dayB, previous).Scan(&total); err != nil || total != 3 {
			t.Fatal("date slots or previous roster changed", total, err)
		}
		_, err = service.ReviewSwap(ctx, "102", SwapRequest{ID: request.ID, Action: "approve"})
		code(err, 409)
	})
	t.Run("schedule edit invalidates pending approvals", func(t *testing.T) {
		value, err := service.CreateSwap(ctx, "301", SwapRequest{DateA: dayA, DateB: dayB})
		if err != nil {
			t.Fatal(err)
		}
		request := value.(Swap)
		if _, err = service.ReviewSwap(ctx, "201", SwapRequest{ID: request.ID, Action: "claim"}); err != nil {
			t.Fatal(err)
		}
		_, err = service.Run(ctx, "101", func(w *work) (any, error) { return w.day(dayA, false) })
		if err != nil {
			t.Fatal(err)
		}
		current, err := service.Run(ctx, "101", func(w *work) (any, error) { return w.day(dayA, false) })
		if err != nil {
			t.Fatal(err)
		}
		a = current.(Day)
		a = change("101", "extra", Request{Date: dayA, Revision: a.Revision, Wid: "103", Zone: "FX", Reason: "edited"})
		_, err = service.ReviewSwap(ctx, "102", SwapRequest{ID: request.ID, Action: "approve"})
		code(err, 409)
		var status string
		if err = admin.QueryRow(ctx, `SELECT status FROM scheduler.swap_requests WHERE id=$1`, request.ID).Scan(&status); err != nil || status != "stale" {
			t.Fatal(status, err)
		}
	})
	t.Run("current role overrides old login", func(t *testing.T) {
		if _, err = admin.Exec(ctx, `UPDATE wts.operators SET access='pre-member' WHERE wid='101'`); err != nil {
			t.Fatal(err)
		}
		_, err = service.Change(ctx, "101", "extra", Request{Date: dayA, Revision: a.Revision, Wid: "102", Zone: "FX", Reason: "old token"})
		code(err, 403)
		if _, err = admin.Exec(ctx, `UPDATE wts.operators SET access='group-leader' WHERE wid='101'`); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("availability cannot mutate other members", func(t *testing.T) {
		yes := true
		if _, err = service.Availability(ctx, "201", AvailabilityRequest{FreeDays: []int{1, 3}, Date: dayA, Available: &yes}); err != nil {
			t.Fatal(err)
		}
		value, err := service.Overview(ctx, "201", dayA[:7])
		if err != nil {
			t.Fatal(err)
		}
		if len(value.(Overview).FreeDays) != 2 {
			t.Fatal("availability not saved")
		}
		var others int
		err = admin.QueryRow(ctx, `SELECT count(*) FROM scheduler.freeday WHERE wid<>'201'`).Scan(&others)
		if err != nil || others != 0 {
			t.Fatal("other availability changed")
		}
		if _, err = service.Audit(ctx, "101", dayA); err != nil {
			t.Fatal(err)
		}
		_, err = service.Audit(ctx, "201", dayA)
		code(err, 403)
	})
	t.Run("concurrent edits cannot overwrite each other", func(t *testing.T) {
		var wg sync.WaitGroup
		errorsFound := make(chan error, 2)
		for _, wid := range []string{"201", "202"} {
			wg.Add(1)
			go func(target string) {
				defer wg.Done()
				_, failure := service.Change(ctx, "103", "extra", Request{Date: dayC, Revision: c.Revision, Wid: target, Zone: "FX", Reason: "concurrent test"})
				errorsFound <- failure
			}(wid)
		}
		wg.Wait()
		close(errorsFound)
		passed, conflicted := 0, 0
		for failure := range errorsFound {
			if failure == nil {
				passed++
			} else {
				code(failure, 409)
				conflicted++
			}
		}
		if passed != 1 || conflicted != 1 {
			t.Fatal("concurrent revision check failed", passed, conflicted)
		}
	})
	t.Run("only one member can claim a hall request", func(t *testing.T) {
		dateD := time.Now().In(china).AddDate(0, 0, 7).Format(time.DateOnly)
		dateE := time.Now().In(china).AddDate(0, 0, 8).Format(time.DateOnly)
		currentRevision := func(date string) int {
			v, e := service.Run(ctx, "900", func(w *work) (any, error) { return w.day(date, false) })
			if e == nil {
				return v.(Day).Revision
			}
			var se *Error
			if errors.As(e, &se) && se.Code == 404 {
				return 0
			}
			t.Fatal(e)
			return 0
		}
		d := change("900", "group", Request{Date: dateD, Revision: currentRevision(dateD), LeaderWid: "101", Wids: []string{"201"}, Reason: "concurrent claim fixture"})
		e := change("900", "group", Request{Date: dateE, Revision: currentRevision(dateE), LeaderWid: "102", Wids: []string{"301", "302"}, Reason: "concurrent claim fixture"})
		published("101", d, []Assignment{{Wid: "101", Zone: "FX"}, {Wid: "201", Zone: "BM"}})
		published("102", e, []Assignment{{Wid: "102", Zone: "FX"}, {Wid: "301", Zone: "BM"}, {Wid: "302", Zone: "FX"}})
		value, failure := service.CreateSwap(ctx, "201", SwapRequest{DateA: dateD, DateB: dateE})
		if failure != nil {
			t.Fatal(failure)
		}
		request := value.(Swap)
		var wg sync.WaitGroup
		results := make(chan error, 2)
		for _, wid := range []string{"301", "302"} {
			wg.Add(1)
			go func(actor string) {
				defer wg.Done()
				_, failure := service.ReviewSwap(ctx, actor, SwapRequest{ID: request.ID, Action: "claim"})
				results <- failure
			}(wid)
		}
		wg.Wait()
		close(results)
		accepted, conflicted := 0, 0
		for failure := range results {
			if failure == nil {
				accepted++
			} else {
				code(failure, 409)
				conflicted++
			}
		}
		if accepted != 1 || conflicted != 1 {
			t.Fatal("request claimed more than once", accepted, conflicted)
		}
		var state, target string
		if failure = admin.QueryRow(ctx, `SELECT status,target_wid FROM scheduler.swap_requests WHERE id=$1`, request.ID).Scan(&state, &target); failure != nil {
			t.Fatal(failure)
		}
		if state != "pending" || (target != "301" && target != "302") {
			t.Fatal("claim did not enter approval", state, target)
		}
		if _, failure = service.ReviewSwap(ctx, "201", SwapRequest{ID: request.ID, Action: "cancel"}); failure != nil {
			t.Fatal(failure)
		}
	})
}
