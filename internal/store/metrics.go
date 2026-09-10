package store

import (
	"fmt"
	"time"
)

// PeriodMetrics is computed from current records; it is never persisted.
type PeriodMetrics struct {
	Period             Period
	FocusTime          time.Duration
	Sessions           int
	Tasks              []TaskTotal
	CompletedTodos     int
	HabitCompletions   int
	HabitOpportunities int
}

// HabitPercent returns 0 when no habit-days are eligible.
func (m PeriodMetrics) HabitPercent() float64 {
	if m.HabitOpportunities == 0 {
		return 0
	}
	return 100 * float64(m.HabitCompletions) / float64(m.HabitOpportunities)
}

// Metrics counts positive session overlaps and completed todos in [Start, End).
// Active sessions and future timestamps are capped at now. A crossing session
// can count in both adjacent periods, but its time is split at the boundary.
func (db *Database) Metrics(period Period, now time.Time) (PeriodMetrics, error) {
	result := PeriodMetrics{Period: period}
	canonical, err := NewPeriod(period.Kind, period.Start)
	if err != nil {
		return result, err
	}
	if !canonical.Start.Equal(period.Start) || !canonical.End.Equal(period.End) {
		return result, fmt.Errorf("limites de periodo invalidos")
	}
	if period.Start.After(now) {
		return result, nil
	}
	end := period.End
	if now.Before(end) {
		end = now
	}
	result.Tasks = TotalsWithin(db.Entries, period.Start, end)
	for _, task := range result.Tasks {
		result.FocusTime += task.Duration
		result.Sessions += task.Sessions
	}
	for _, todo := range db.Todos {
		if todo.CompletedAt != nil && !todo.CompletedAt.Before(period.Start) &&
			todo.CompletedAt.Before(period.End) && !todo.CompletedAt.After(now) {
			result.CompletedTodos++
		}
	}
	loc := period.Start.Location()
	today := StartOfDay(now.In(loc))
	for _, habit := range db.Habits {
		if habit.CreatedAt.After(now) {
			continue
		}
		first := StartOfDay(habit.CreatedAt.In(loc))
		if first.Before(period.Start) {
			first = period.Start
		}
		// Membership removes duplicate marks and ignores malformed/out-of-range dates.
		marks := make(map[string]bool, len(habit.Completions))
		for _, date := range habit.Completions {
			marks[date] = true
		}
		for day := first; day.Before(period.End) && !day.After(today); day = day.AddDate(0, 0, 1) {
			result.HabitOpportunities++
			if marks[day.Format("2006-01-02")] {
				result.HabitCompletions++
			}
		}
	}
	return result, nil
}
