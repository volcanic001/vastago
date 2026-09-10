package store

import (
	"fmt"
	"time"
)

type PeriodKind string

const (
	Day   PeriodKind = "day"
	Week  PeriodKind = "week"
	Month PeriodKind = "month"
	Year  PeriodKind = "year"
)

// Period is a calendar interval [Start, End) in Start's location.
type Period struct {
	Kind  PeriodKind
	Start time.Time
	End   time.Time
}

// NewPeriod uses the reference's location. Weeks begin on Monday.
func NewPeriod(kind PeriodKind, reference time.Time) (Period, error) {
	start := StartOfDay(reference)
	var end time.Time
	switch kind {
	case Day:
		end = start.AddDate(0, 0, 1)
	case Week:
		start = start.AddDate(0, 0, -(int(start.Weekday())+6)%7)
		end = start.AddDate(0, 0, 7)
	case Month:
		start = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
		end = start.AddDate(0, 1, 0)
	case Year:
		start = time.Date(start.Year(), time.January, 1, 0, 0, 0, 0, start.Location())
		end = start.AddDate(1, 0, 0)
	default:
		return Period{}, fmt.Errorf("periodo desconocido: %q", kind)
	}
	return Period{Kind: kind, Start: start, End: end}, nil
}

// Shift moves whole calendar periods, avoiding month-end and DST drift.
func (p Period) Shift(offset int) (Period, error) {
	switch p.Kind {
	case Day:
		return NewPeriod(p.Kind, p.Start.AddDate(0, 0, offset))
	case Week:
		return NewPeriod(p.Kind, p.Start.AddDate(0, 0, 7*offset))
	case Month:
		return NewPeriod(p.Kind, p.Start.AddDate(0, offset, 0))
	case Year:
		return NewPeriod(p.Kind, p.Start.AddDate(offset, 0, 0))
	default:
		return Period{}, fmt.Errorf("periodo desconocido: %q", p.Kind)
	}
}
