package store

import (
	"testing"
	"time"
)

func TestActivityIncludesCalendarDaysAndSplitsMidnight(t *testing.T) {
	loc := time.FixedZone("local", -6*3600)
	start := time.Date(2026, 1, 1, 23, 0, 0, 0, loc)
	end := start.Add(2 * time.Hour)
	db := &Database{Entries: []Entry{{Task: "Cruza", Start: start, End: &end}}}
	p, _ := NewPeriod(Month, start)
	activity, err := db.Activity(p, time.Date(2026, 1, 2, 2, 0, 0, 0, loc))
	if err != nil || len(activity) != 31 {
		t.Fatalf("activity = %#v, %v", activity, err)
	}
	if activity[0].Duration != time.Hour || activity[1].Duration != time.Hour {
		t.Fatalf("split = %v, %v", activity[0].Duration, activity[1].Duration)
	}
	if activity[2].Duration != 0 || activity[30].Duration != 0 {
		t.Fatal("future calendar days must be zero")
	}
}

func TestActivityRejectsNonCanonicalPeriod(t *testing.T) {
	p, _ := NewPeriod(Day, time.Now())
	p.End = p.End.Add(time.Hour)
	if _, err := (&Database{}).Activity(p, time.Now()); err == nil {
		t.Fatal("accepted invalid period")
	}
}
