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

func habitCalendarLayout(width int) (labelWidth, gap int, spaced, visible bool) {
	if width < 14 {
		return width, 0, false, false
	}
	spaced = width >= compactBreakpoint
	calendarWidth := 7
	gap = 1
	if spaced {
		calendarWidth = 13
		gap = 2
	}
	return width - calendarWidth - gap, gap, spaced, true
}

func habitListLine(habit store.Habit, now time.Time, selected, completed bool, width int) string {
	labelWidth, gap, spaced, visible := habitCalendarLayout(width)
	label := "○ " + trimToWidth(habit.Name, max(1, labelWidth-2))
	if completed {
		label = "● " + trimToWidth(habit.Name, max(1, labelWidth-2))
	}
	if selected {
		label = selectionStyle.Width(labelWidth).Render(label)
	} else if completed {
		label = mutedStyle.Width(labelWidth).Render(label)
	} else {
		label = valueStyle.Width(labelWidth).Render(label)
	}
	if !visible {
		return label
	}
	return label + strings.Repeat(" ", gap) + habitWeekGrid(habit, now, spaced)
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
