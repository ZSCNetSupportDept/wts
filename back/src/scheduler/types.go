package scheduler

import (
	"fmt"
	"time"
)

type Person struct {
	Wid    string `json:"wid"`
	Name   string `json:"name"`
	Access string `json:"access"`
}

type Member struct {
	Person
	Zone       string `json:"zone"`
	AreaLeader bool   `json:"area_leader"`
	Extra      bool   `json:"extra"`
	Available  bool   `json:"available"`
}

type Day struct {
	Date      string   `json:"date"`
	Leader    Person   `json:"leader"`
	Published bool     `json:"published"`
	Revision  int      `json:"revision"`
	Members   []Member `json:"members"`
}

type Exception struct {
	Date      string `json:"date"`
	Available bool   `json:"available"`
}

type Overview struct {
	Me         Person      `json:"me"`
	Manager    bool        `json:"manager"`
	Days       []Day       `json:"days"`
	FreeDays   []int       `json:"free_days"`
	Exceptions []Exception `json:"exceptions"`
}

type Assignment struct {
	Wid        string `json:"wid"`
	Zone       string `json:"zone"`
	AreaLeader bool   `json:"area_leader"`
}

// A single date is a 16:30–18:00 shift. Scope comes from the database, never JWT roles.
func isManager(role string) bool { return role == "chief" || role == "dev" || role == "api" }
func isOperator(role string) bool {
	switch role {
	case "chief", "dev", "api", "group-leader", "formal-member", "informal-member":
		return true
	}
	return false
}
func canArrange(actor Person, day Day) bool {
	return isManager(actor.Access) || actor.Access == "group-leader" && actor.Wid == day.Leader.Wid
}
func validZone(zone string) bool {
	switch zone {
	case "FX", "BM", "DM", "QT", "XHAB", "XHCD", "ZH", "other":
		return true
	}
	return false
}

var china = time.FixedZone("Asia/Shanghai", 8*60*60)

func parseDate(value string) (time.Time, error) {
	d, err := time.ParseInLocation(time.DateOnly, value, china)
	if err != nil || d.Format(time.DateOnly) != value {
		return time.Time{}, fail(400, "日期必须为 YYYY-MM-DD")
	}
	return d, nil
}

type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string            { return e.Message }
func fail(code int, message string) error { return &Error{code, message} }
func checkRevision(day Day, revision int) error {
	if day.Revision != revision {
		return fail(409, "排班已被他人修改，请刷新后重试")
	}
	return nil
}
func validateAssignments(day Day, assignments []Assignment) error {
	expected := map[string]bool{}
	for _, m := range day.Members {
		if !m.Extra {
			expected[m.Wid] = true
		}
	}
	if len(assignments) != len(expected) {
		return fail(400, "必须为当天全部小组成员安排片区")
	}
	seen := map[string]bool{}
	leaders := map[string]bool{}
	for _, a := range assignments {
		if !expected[a.Wid] || seen[a.Wid] {
			return fail(400, "成员不属于当天小组或被重复安排")
		}
		if !validZone(a.Zone) {
			return fail(400, "请选择有效片区")
		}
		if a.AreaLeader && leaders[a.Zone] {
			return fail(400, fmt.Sprintf("片区 %s 只能有一位负责人", a.Zone))
		}
		if a.AreaLeader {
			leaders[a.Zone] = true
		}
		seen[a.Wid] = true
	}
	return nil
}
