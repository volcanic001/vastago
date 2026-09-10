package store

import (
	"testing"
	"time"
	_ "time/tzdata"
)

func TestCalendarPeriods(t *testing.T) {
	loc := time.FixedZone("El Salvador", -6*3600)
	for _, tc := range []struct {
		kind            PeriodKind
		ref, start, end string
	}{
		{Day, "2024-02-29", "2024-02-29", "2024-03-01"},
		{Week, "2023-01-01", "2022-12-26", "2023-01-02"},
		{Week, "2026-09-07", "2026-09-07", "2026-09-14"},
		{Month, "2024-02-29", "2024-02-01", "2024-03-01"},
		{Year, "2024-12-31", "2024-01-01", "2025-01-01"},
	} {
		ref, _ := time.ParseInLocation("2006-01-02", tc.ref, loc)
		p, err := NewPeriod(tc.kind, ref.Add(15*time.Hour))
		if err != nil {
			t.Fatal(err)
		}
		if p.Start.Format("2006-01-02") != tc.start || p.End.Format("2006-01-02") != tc.end || p.Start.Hour() != 0 || p.Start.Location() != loc {
			t.Fatalf("%s: %#v", tc.ref, p)
		}
		next, err := p.Shift(1)
		if err != nil || !next.Start.Equal(p.End) {
			t.Fatalf("shift: %#v %v", next, err)
		}
		previous, err := next.Shift(-1)
		if err != nil || !previous.Start.Equal(p.Start) {
			t.Fatal("shift did not round trip")
		}
	}
	if _, err := NewPeriod("invalid", time.Now()); err == nil {
		t.Fatal("accepted invalid kind")
	}
}

func TestCalendarDaysAcrossDST(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		month      time.Month
		day, hours int
	}{{3, 8, 23}, {11, 1, 25}} {
		p, err := NewPeriod(Day, time.Date(2026, tc.month, tc.day, 12, 0, 0, 0, loc))
		if err != nil {
			t.Fatal(err)
		}
		if p.End.Sub(p.Start) != time.Duration(tc.hours)*time.Hour {
			t.Fatalf("wrong DST interval: %#v", p)
		}
	}
}
