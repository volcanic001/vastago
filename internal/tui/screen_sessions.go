package tui

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/volcanic001/vastago/internal/store"
	"strings"
)

func (m Model) history(width int) string {
	narrow := width < compactBreakpoint || m.height < 26
	limit := max(1, m.height-8)
	if narrow {
		limit = max(1, (m.height-7)/2)
	}
	if limit > 14 {
		limit = 14
	}
	entries := m.db.Recent(limit)
	if len(entries) == 0 {
		if narrow {
			return "\n" + mutedStyle.Render(trimToWidth("Aun no hay sesiones. Pulsa n para comenzar.", width))
		}
		return "\n" + panelStyle.Width(max(20, width-4)).Render(mutedStyle.Render("Aun no hay sesiones. Pulsa n para comenzar."))
	}
	lines := []string{mutedStyle.Render("HISTORIAL RECIENTE")}
	for _, entry := range entries {
		state := store.FormatDuration(entry.Duration(m.now))
		if entry.End == nil {
			state += " · activa"
		}
		if width < 26 {
			lines = append(lines, fmt.Sprintf("%s %s", entry.Start.Format("02/01"), trimToWidth(entry.Task, max(3, width-6))))
			lines = append(lines, "  "+mutedStyle.Render(trimToWidth(state, max(3, width-2))))
		} else if narrow {
			lines = append(lines, fmt.Sprintf("%s  %s", entry.Start.Format("02/01 15:04"), trimToWidth(entry.Task, max(6, width-15))))
			lines = append(lines, "           "+mutedStyle.Render(state))
		} else {
			date := mutedStyle.Render(entry.Start.Format("Mon 02 · 15:04"))
			available := max(8, width-lipgloss.Width(date)-lipgloss.Width(state)-8)
			lines = append(lines, date+"  "+valueStyle.Render(trimToWidth(entry.Task, available))+"  "+mutedStyle.Render(state))
		}
	}
	if narrow {
		return "\n" + strings.Join(lines, "\n")
	}
	return "\n" + panelStyle.Width(max(20, width-4)).Render(strings.Join(lines, "\n"))
}
