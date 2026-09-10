package store

import (
	"fmt"
	"time"
)

// DayActivity is the focused time assigned to one local calendar day.
type DayActivity struct {
	Date     time.Time
	Duration time.Duration
}

// Activity returns every day in period, including future days as zeroes. Time
// is capped at now and split at each local midnight.
func (db *Database) Activity(period Period, now time.Time) ([]DayActivity, error) {
	canonical, err := NewPeriod(period.Kind, period.Start)
	if err != nil {
		return nil, err
	}
	if !canonical.Start.Equal(period.Start) || !canonical.End.Equal(period.End) {
		return nil, fmt.Errorf("limites de periodo invalidos")
	}
	result := make([]DayActivity, 0, 32)
	for day := period.Start; day.Before(period.End); day = day.AddDate(0, 0, 1) {
		dayEnd := day.AddDate(0, 0, 1)
		end := dayEnd
		if now.Before(end) {
			end = now
		}
		var duration time.Duration
		if end.After(day) {
			for _, entry := range db.Entries {
				duration += entry.DurationWithin(day, end)
			}
		}
		result = append(result, DayActivity{Date: day, Duration: duration})
	}
	return result, nil
}
