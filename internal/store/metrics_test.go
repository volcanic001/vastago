package store

import (
	"testing"
	"time"
)

func TestMetricsSplitSessionsAtMidnightAndCapAtNow(t *testing.T) {
	loc := time.FixedZone("local", -6*3600)
	midnight := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
	now := midnight.Add(2 * time.Hour)
	end := midnight.Add(time.Hour)
	boundary := midnight
	future := now.Add(time.Hour)
	db := &Database{Entries: []Entry{
		{Task: "Cross", Start: midnight.Add(-time.Hour), End: &end},
		{Task: "Active", Start: midnight.Add(time.Hour)},
		{Task: "Boundary", Start: midnight.Add(-time.Hour), End: &boundary},
		{Task: "Future", Start: future},
		{Task: "Invalid", Start: future, End: &boundary},
		{Task: "Zero", Start: midnight, End: &boundary},
	}}
	p, _ := NewPeriod(Day, midnight)
	got, err := db.Metrics(p, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.FocusTime != 2*time.Hour || got.Sessions != 2 {
		t.Fatalf("today: %+v", got)
	}
	prior, _ := p.Shift(-1)
	got, err = db.Metrics(prior, now)
	if err != nil {
		t.Fatal(err)
	}
	if got.FocusTime != 2*time.Hour || got.Sessions != 2 {
		t.Fatalf("prior: %+v", got)
	}
	for _, kind := range []PeriodKind{Month, Year} {
		period, _ := NewPeriod(kind, midnight)
		got, err = db.Metrics(period, now)
		if err != nil || got.FocusTime != 2*time.Hour {
			t.Fatalf("%s: %+v %v", kind, got, err)
		}
	}
}

func TestMetricsTodosAndHabitCalendarEligibility(t *testing.T) {
	loc := time.FixedZone("local", -6*3600)
	start := time.Date(2026, 9, 7, 0, 0, 0, 0, loc) // Monday
	now := start.AddDate(0, 0, 2).Add(12 * time.Hour)
	atStart := start
	atEnd := start.AddDate(0, 0, 7)
	before := start.Add(-time.Second)
	future := now.Add(time.Hour)
	db := &Database{
		Todos: []Todo{{CompletedAt: &atStart}, {CompletedAt: &now}, {CompletedAt: &before}, {CompletedAt: &atEnd}, {CompletedAt: &future}, {}},
		Habits: []Habit{
			{CreatedAt: start.AddDate(0, 0, -2), Completions: []string{"2026-09-07", "2026-09-07", "2026-09-08", "2026-09-10", "invalid"}},
			// UTC creation falls on Tuesday locally.
			{CreatedAt: time.Date(2026, 9, 9, 1, 0, 0, 0, time.UTC), Completions: []string{"2026-09-07", "2026-09-09"}},
			{CreatedAt: future, Completions: []string{"2026-09-09"}},
		},
	}
	p, _ := NewPeriod(Week, now)
	got, err := db.Metrics(p, now)
	if err != nil {
		t.Fatal(err)
	}
	// Three days for the first habit, two for the second. Today is included.
	if got.CompletedTodos != 2 || got.HabitOpportunities != 5 || got.HabitCompletions != 3 || got.HabitPercent() != 60 {
		t.Fatalf("metrics: %+v percent %v", got, got.HabitPercent())
	}
	next, _ := p.Shift(1)
	got, err = db.Metrics(next, now)
	if err != nil || got.HabitOpportunities != 0 || got.CompletedTodos != 0 || got.HabitPercent() != 0 {
		t.Fatalf("future: %+v %v", got, err)
	}
}

func TestMetricsLeapYearAndDSTDays(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)
	p, _ := NewPeriod(Year, start)
	db := &Database{Habits: []Habit{{CreatedAt: start}}}
	got, err := db.Metrics(p, p.End)
	if err != nil || got.HabitOpportunities != 366 {
		t.Fatalf("year: %+v %v", got, err)
	}
	p, _ = NewPeriod(Day, time.Date(2026, 3, 8, 0, 0, 0, 0, loc))
	db.Entries = []Entry{{Task: "DST", Start: p.Start, End: &p.End}}
	got, err = db.Metrics(p, p.End)
	if err != nil || got.FocusTime != 23*time.Hour || got.HabitOpportunities != 1 {
		t.Fatalf("DST: %+v %v", got, err)
	}
}

func TestMetricsEmptyAndInvalidPeriod(t *testing.T) {
	db := &Database{}
	p, _ := NewPeriod(Month, time.Now())
	got, err := db.Metrics(p, time.Now())
	if err != nil || got.FocusTime != 0 || got.Sessions != 0 || got.HabitPercent() != 0 {
		t.Fatalf("%+v %v", got, err)
	}
	p.End = p.Start
	if _, err := db.Metrics(p, time.Now()); err == nil {
		t.Fatal("accepted invalid interval")
	}
}
