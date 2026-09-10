package tui

import (
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

func weekStart(now time.Time) time.Time {
	day := store.StartOfDay(now)
	return day.AddDate(0, 0, -(int(day.Weekday())+6)%7)
}

func habitWeekGrid(habit store.Habit, now time.Time, spaced bool) string {
	marks := make(map[string]bool, len(habit.Completions))
	for _, completion := range habit.Completions {
		marks[completion] = true
	}
	start := weekStart(now)
	parts := make([]string, 0, 7)
	for offset := 0; offset < 7; offset++ {
		day := start.AddDate(0, 0, offset)
		mark := mutedStyle.Render("·")
		if marks[day.Format("2006-01-02")] {
			mark = lipgloss.NewStyle().Foreground(colorLeaf).Render("●")
		}
		parts = append(parts, mark)
	}
	separator := ""
	if spaced {
		separator = " "
	}
	return strings.Join(parts, separator)
}
