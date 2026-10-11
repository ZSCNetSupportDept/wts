package scheduler

import "testing"

func TestDateScope(t *testing.T) {
	day := Day{Leader: Person{Wid: "101"}}
	for _, test := range []struct {
		actor   Person
		allowed bool
	}{
		{Person{Wid: "900", Access: "chief"}, true}, {Person{Access: "dev"}, true}, {Person{Access: "api"}, true},
		{Person{Wid: "101", Access: "group-leader"}, true}, {Person{Wid: "102", Access: "group-leader"}, false},
		{Person{Wid: "101", Access: "formal-member"}, false}, {Person{Access: "user"}, false}, {Person{Access: "pre-member"}, false},
	} {
		if canArrange(test.actor, day) != test.allowed {
			t.Errorf("unexpected permission: %+v", test.actor)
		}
	}
}
func TestRandomPreservesCapacitiesAndFixedMembers(t *testing.T) {
	people := map[string]Member{"l": {Person: Person{Access: "group-leader"}}, "f": {Person: Person{Access: "formal-member"}},
		"i1": {Person: Person{Access: "informal-member"}}, "i2": {Person: Person{Access: "informal-member"}}}
	input := []Assignment{{Wid: "l", Zone: "FX", AreaLeader: true}, {Wid: "f", Zone: "BM"}, {Wid: "i1", Zone: "FX"}, {Wid: "i2", Zone: "BM"}}
	for n := 0; n < 30; n++ {
		result, err := randomize(input, people, map[string]int{"FX": 2, "BM": 2})
		if err != nil {
			t.Fatal(err)
		}
		if result[0] != input[0] || result[1] != input[1] {
			t.Fatal("fixed members were moved")
		}
		counts := map[string]int{}
		for _, a := range result {
			counts[a.Zone]++
		}
		if counts["FX"] != 2 || counts["BM"] != 2 {
			t.Fatal(counts)
		}
	}
	if _, err := randomize(input, people, map[string]int{"FX": 1, "BM": 1}); err == nil {
		t.Fatal("must reject incompatible group size")
	}
}
func TestAssignmentsCannotAddForeignMembers(t *testing.T) {
	day := Day{Members: []Member{{Person: Person{Wid: "1"}}, {Person: Person{Wid: "2"}}}}
	if err := validateAssignments(day, []Assignment{{Wid: "1", Zone: "FX"}, {Wid: "x", Zone: "BM"}}); err == nil {
		t.Fatal("foreign member accepted")
	}
	if err := validateAssignments(day, []Assignment{{Wid: "1", Zone: "FX", AreaLeader: true}, {Wid: "2", Zone: "FX", AreaLeader: true}}); err == nil {
		t.Fatal("duplicate area leaders accepted")
	}
}
