package tui

import (
	"fmt"
	"strings"
)

func (m Model) todosScreen(width int) string {
	if len(m.db.Todos) == 0 {
		return m.emptyScreen(width, "PENDIENTES", "Aun no hay pendientes. Pulsa n para crear uno.")
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
	rows := todoRows(m.db.Todos)
	selectedRow := 0
	for row, item := range rows {
		if item.todoIndex == m.selectedTodo {
			selectedRow = row
			break
		}
	}
	start := listWindowStart(selectedRow, len(rows), limit)
	end := min(len(rows), start+limit)

	summary := "PENDIENTES"
	lines := []string{mutedStyle.Render(trimToWidth(summary, contentWidth))}
	for _, row := range rows[start:end] {
		if row.todoIndex < 0 {
			lines = append(lines, mutedStyle.Render(trimToWidth(row.heading, contentWidth)))
			continue
		}
		todo := m.db.Todos[row.todoIndex]
		marker := "○"
		if todo.Completed() {
			marker = "●"
		}
		line := fmt.Sprintf("%s %s", marker, trimToWidth(todo.Title, max(1, contentWidth-2)))
		if row.todoIndex == m.selectedTodo {
			line = selectionStyle.Width(contentWidth).Render(line)
		} else if todo.Completed() {
			line = mutedStyle.Render(line)
		} else {
			line = valueStyle.Render(line)
		}
		lines = append(lines, line)
	}
	if start > 0 {
		lines[0] += mutedStyle.Render(" · ↑")
	}
	if end < len(rows) {
		lines[0] += mutedStyle.Render(" · ↓")
	}

	return "\n" + strings.Join(lines, "\n")
}

func listWindowStart(selected, total, limit int) int {
	if total <= limit || selected < limit {
		return 0
	}
	start := selected - limit + 1
	if start > total-limit {
		return total - limit
	}
	return start
}
