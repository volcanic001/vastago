package tui

import (
	"charm.land/lipgloss/v2"
	"github.com/volcanic001/vastago/internal/store"
)

func (m Model) dashboard(width int) string {
	if m.width < compactBreakpoint || m.height < 26 {
		return m.tinyDashboard(width)
	}
	if m.width < wideBreakpoint || m.height < 24 {
		return m.compactDashboard(width)
	}
	leftWidth := width*5/9 - 1
	rightWidth := width - leftWidth - 2
	left := panelStyle.Width(leftWidth - 4).Render(m.activeContent(leftWidth - 4))
	right := panelStyle.Width(rightWidth - 4).Render(m.statsContent(rightWidth-4, 5))
	return "\n" + lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

func (m Model) compactDashboard(width int) string {
	contentWidth := max(20, width-4)
	active := panelStyle.Width(contentWidth).Render(m.activeContent(contentWidth))
	stats := panelStyle.Width(contentWidth).Render(m.statsContent(contentWidth, 3))
	return "\n" + lipgloss.JoinVertical(lipgloss.Left, active, stats)
}

func (m Model) tinyDashboard(width int) string {
	active := m.db.Active()
	line := goldStyle.Render(trimToWidth("SIN SESIÓN ACTIVA", width))
	if active != nil {
		line = titleStyle.Render(trimToWidth(active.Task, width-12)) + "  " + valueStyle.Render(store.FormatDuration(active.Duration(m.now)))
	}
	today := m.totalSince(store.StartOfDay(m.now))
	week := m.totalSince(store.StartOfDay(m.now).AddDate(0, 0, -6))
	if width < 30 {
		return "\n" + line + "\n" + mutedStyle.Render("hoy ") + valueStyle.Render(store.FormatDuration(today)) +
			"\n" + mutedStyle.Render("7 dias ") + valueStyle.Render(store.FormatDuration(week))
	}
	return "\n" + line + "\n" + mutedStyle.Render("hoy ") + valueStyle.Render(store.FormatDuration(today)) +
		"  " + mutedStyle.Render("7 dias ") + valueStyle.Render(store.FormatDuration(week))
}

func (m Model) activeContent(width int) string {
	active := m.db.Active()
	if active == nil {
		return mutedStyle.Render("AHORA") + "\n\n" + goldStyle.Render("Sin sesión activa") + "\n" + mutedStyle.Render("Pulsa n para iniciar una sesión.")
	}
	task := valueStyle.Render(trimToWidth(active.Task, width))
	duration := titleStyle.Render(store.FormatDuration(active.Duration(m.now)))
	started := mutedStyle.Render("desde " + active.Start.Format("15:04"))
	result := mutedStyle.Render("AHORA") + "\n\n" + task + "\n" + duration + "  " + started
	if active.Note != "" {
		result += "\n" + mutedStyle.Render(trimToWidth(active.Note, width))
	}
	return result
}
