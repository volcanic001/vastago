package tui

import (
	"fmt"
	"strings"
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

	summary := fmt.Sprintf("HABITOS · %d/%d hoy", completed, len(m.db.Habits))
	lines := []string{mutedStyle.Render(trimToWidth(summary, contentWidth))}
	for index := start; index < end; index++ {
		habit := m.db.Habits[index]
		marker := "○"
		if habit.CompletedOn(m.now) {
			marker = "●"
		}
		line := fmt.Sprintf("%s %s", marker, trimToWidth(habit.Name, max(1, contentWidth-2)))
		if index == m.selectedHabit {
			line = selectionStyle.Width(contentWidth).Render(line)
		} else if habit.CompletedOn(m.now) {
			line = mutedStyle.Render(line)
		} else {
			line = valueStyle.Render(line)
		}
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
