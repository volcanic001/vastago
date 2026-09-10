package tui

import (
	"fmt"
	"time"

	"github.com/volcanic001/vastago/internal/store"
)

func (m Model) statsContent(width, taskLimit int) string {
	todayStart := store.StartOfDay(m.now)
	weekStart := todayStart.AddDate(0, 0, -6)
	entries := m.db.Since(weekStart, m.now)
	totals := store.TotalsWithin(entries, weekStart, m.now)
	result := mutedStyle.Render("RITMO") + "\n\n"
	result += statLine("Hoy", store.FormatDuration(m.totalSince(todayStart)), width) + "\n"
	result += statLine("Semana", store.FormatDuration(m.totalSince(weekStart)), width)
	habitsDone, habitsTotal := m.todayHabits()
	result += "\n\n" + mutedStyle.Render("HOY")
	result += "\n" + statLine("Habitos", fmt.Sprintf("%d/%d hoy", habitsDone, habitsTotal), width)
	result += "\n" + statLine("Pendientes", fmt.Sprintf("%d abiertos", m.openTodos()), width)
	if len(totals) > 0 {
		result += "\n\n" + mutedStyle.Render("MAS CULTIVADO")
	}
	for index, total := range totals {
		if index >= taskLimit {
			break
		}
		result += "\n" + statLine(trimToWidth(total.Task, width/2), store.FormatDuration(total.Duration), width)
	}
	return result
}

func (m Model) todayHabits() (completed, total int) {
	for _, habit := range m.db.Habits {
		total++
		if habit.CompletedOn(m.now) {
			completed++
		}
	}
	return completed, total
}

func (m Model) openTodos() int {
	open := 0
	for _, todo := range m.db.Todos {
		if !todo.Completed() {
			open++
		}
	}
	return open
}

func (m Model) totalSince(start time.Time) time.Duration {
	var total time.Duration
	for _, entry := range m.db.Since(start, m.now) {
		total += entry.DurationWithin(start, m.now)
	}
	return total
}
