package tui

import (
	"fmt"
	"strings"

	"github.com/volcanic001/vastago/internal/store"
)

func (m Model) habitsScreen(width int) string {
	if len(m.db.Habits) == 0 {
		return m.emptyScreen(width, "HABITOS", "Aun no hay habitos. Pulsa n para crear uno.")
	}

	completed := 0
	for _, habit := range m.db.Habits {
		if habit.CompletedOn(m.now) {
			completed++
		}
	}
	week, _ := store.NewPeriod(store.Week, m.now)
	weekly, _ := m.db.Metrics(week, m.now)
	contentWidth := width
	narrow := m.width < compactBreakpoint || m.height < 26
	limit := max(1, m.height-8)
	if narrow {
		limit = max(1, m.height-7)
	}
	if limit > 16 {
		limit = 16
	}
	start := listWindowStart(m.selectedHabit, len(m.db.Habits), limit)
	end := min(len(m.db.Habits), start+limit)

	summary := fmt.Sprintf("HABITOS · %d/%d hoy · %d/%d semana", completed, len(m.db.Habits), weekly.HabitCompletions, weekly.HabitOpportunities)
	labelWidth, gap, showWeekHeader, _ := habitCalendarLayout(contentWidth)
	lines := []string{mutedStyle.Render(trimToWidth(summary, contentWidth))}
	if showWeekHeader {
		lines = append(lines, mutedStyle.Render(strings.Repeat(" ", labelWidth+gap)+"L M X J V S D"))
	}
	for index := start; index < end; index++ {
		habit := m.db.Habits[index]
		line := habitListLine(habit, m.now, index == m.selectedHabit, habit.CompletedOn(m.now), contentWidth)
		lines = append(lines, line)
	}
	if start > 0 {
		lines[0] += mutedStyle.Render(" · ↑")
	}
	if end < len(m.db.Habits) {
		lines[0] += mutedStyle.Render(" · ↓")
	}

	content := strings.Join(lines, "\n")
	return "\n" + content
}
