package tui

import (
	"fmt"
	"strings"
)

func (m Model) todosScreen(width int) string {
	if len(m.db.Todos) == 0 {
		return m.emptyScreen(width, "PENDIENTES", "Aun no hay pendientes. Pulsa n para crear uno.")
	}

	open := 0
	for _, todo := range m.db.Todos {
		if !todo.Completed() {
			open++
		}
	}
	contentWidth := max(12, width-4)
	narrow := m.width < compactBreakpoint || m.height < 26
	if narrow {
		contentWidth = width
	}
	limit := max(1, m.height-8)
	if narrow {
		limit = max(1, m.height-7)
	}
	if limit > 16 {
		limit = 16
	}
	start := todoWindowStart(m.selectedTodo, len(m.db.Todos), limit)
	end := min(len(m.db.Todos), start+limit)

	lines := []string{mutedStyle.Render(trimToWidth(fmt.Sprintf("PENDIENTES · %d abiertos · %d total", open, len(m.db.Todos)), contentWidth))}
	for index := start; index < end; index++ {
		todo := m.db.Todos[index]
		marker := "○"
		if todo.Completed() {
			marker = "●"
		}
		line := fmt.Sprintf("%s %s", marker, trimToWidth(todo.Title, max(1, contentWidth-2)))
		if index == m.selectedTodo {
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
	if end < len(m.db.Todos) {
		lines[0] += mutedStyle.Render(" · ↓")
	}

	content := strings.Join(lines, "\n")
	if narrow {
		return "\n" + content
	}
	return "\n" + panelStyle.Width(contentWidth).Render(content)
}

func todoWindowStart(selected, total, limit int) int {
	if total <= limit || selected < limit {
		return 0
	}
	start := selected - limit + 1
	if start > total-limit {
		return total - limit
	}
	return start
}
