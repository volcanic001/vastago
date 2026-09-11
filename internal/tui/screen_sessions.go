package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/volcanic001/vastago/internal/store"
)

func (m Model) sessionView(width int) string {
	if m.sessionEdit != nil {
		return m.sessionFormView(width)
	}
	if m.sessionDetailID != "" {
		return m.sessionDetailView(width)
	}
	if len(m.db.Entries) == 0 {
		return m.emptyScreen(width, "SESIONES", "Aun no hay sesiones. Pulsa n para comenzar.")
	}
	selected := max(0, min(m.selectedSession, len(m.db.Entries)-1))
	limit := max(1, m.height-7)
	start := listWindowStart(selected, len(m.db.Entries), limit)
	lines := []string{mutedStyle.Render(trimToWidth(fmt.Sprintf("SESIONES · %d/%d", selected+1, len(m.db.Entries)), width))}
	if m.sessionDeleteID != "" {
		entry := m.db.Entries[len(m.db.Entries)-1-selected]
		label := "Borrar sesion: "
		if entry.End == nil {
			label = "Borrar sesion ACTIVA: "
		}
		return "\n" + errorStyle.Render(trimToWidth(label, width)) + "\n" + valueStyle.Render(trimToWidth(entry.Task, width))
	}
	for row := start; row < min(len(m.db.Entries), start+limit); row++ {
		entry := m.db.Entries[len(m.db.Entries)-1-row]
		line := sessionListLine(entry, m.now, width)
		if row == selected {
			line = selectionStyle.Width(width).Render(line)
		}
		lines = append(lines, line)
	}
	return "\n" + strings.Join(lines, "\n")
}

func sessionListLine(entry store.Entry, now time.Time, width int) string {
	start := entry.Start.Local()
	stamp := start.Format("02/01 15:04")
	if entry.End == nil {
		stamp += "–ahora"
	} else {
		stamp += "–" + entry.End.Local().Format("15:04")
	}
	state := store.FormatDuration(entry.Duration(now))
	if entry.End == nil {
		state += " activa"
	}
	left := fmt.Sprintf("%s · %s", stamp, entry.Task)
	if width < compactBreakpoint {
		left = fmt.Sprintf("%s %s", start.Format("02/01 15:04"), entry.Task)
	}

	stateWidth := lipgloss.Width(state)
	if stateWidth >= width {
		return trimToWidth(state, width)
	}
	left = trimToWidth(left, width-stateWidth-1)
	gap := max(1, width-lipgloss.Width(left)-stateWidth)
	return left + strings.Repeat(" ", gap) + state
}

func (m Model) sessionDetailView(width int) string {
	for _, entry := range m.db.Entries {
		if entry.ID != m.sessionDetailID {
			continue
		}
		end := "activa"
		if entry.End != nil {
			end = entry.End.Local().Format("02/01/2006 15:04")
		}
		lines := []string{titleStyle.Render("DETALLE DE SESION"), valueStyle.Render(trimToWidth(entry.Task, width)), mutedStyle.Render("Inicio " + entry.Start.Local().Format("02/01/2006 15:04")), mutedStyle.Render("Fin " + end), mutedStyle.Render("Duracion " + store.FormatDuration(entry.Duration(m.now)))}
		if entry.Note != "" {
			lines = append(lines, mutedStyle.Render(trimToWidth(entry.Note, width)))
		}
		return "\n" + strings.Join(lines, "\n")
	}
	return m.emptyScreen(width, "SESIONES", "La sesion ya no existe.")
}

func (m Model) sessionFormView(width int) string {
	form := m.sessionEdit
	labels := []string{"Tarea", "Nota", "Inicio", "Fin"}
	lines := []string{titleStyle.Render(trimToWidth("EDITAR SESION", width))}
	count := 4
	if form.active {
		count = 3
	}
	if m.height < 16 || width < 40 {
		lines = append(lines, mutedStyle.Render(trimToWidth(fmt.Sprintf("%d/%d %s", form.field+1, count, labels[form.field]), width)))
		lines = append(lines, valueStyle.Render(trimToWidth(form.values[form.field], width)))
	} else {
		for i := 0; i < count; i++ {
			line := trimToWidth(labels[i]+": "+form.values[i], width)
			if i == form.field {
				line = selectionStyle.Width(width).Render(line)
			}
			lines = append(lines, line)
		}
	}
	return "\n" + strings.Join(lines, "\n")
}

func (m Model) sessionFooter(width int) string {
	items := []shortcut{{key: "j/k/↑↓", action: "mover"}, {key: "i", action: "detalle", primary: true}, {key: "enter/e", action: "editar", primary: true}, {key: "d", action: "borrar", tone: shortcutDanger, primary: true}, {key: "n", action: "nueva"}, {key: "x", action: "fin"}, {key: "tab", action: "vistas"}, {key: "q", action: "salir", primary: true}}
	if m.sessionEdit != nil {
		items = []shortcut{{key: "enter", action: "siguiente/guardar", primary: true}, {key: "tab", action: "campo"}, {key: "ctrl+u", action: "limpiar"}, {key: "esc", action: "cancelar"}}
	}
	if m.sessionDetailID != "" {
		items = []shortcut{{key: "i/esc", action: "volver"}}
	}
	if m.sessionDeleteID != "" {
		items = []shortcut{{key: "y", action: "borrar", tone: shortcutDanger, primary: true}, {key: "n/esc", action: "cancelar"}}
	}
	status := m.message
	if m.err != nil {
		status = "error: " + m.err.Error()
	}
	separator := " · "
	if m.sessionDeleteID != "" {
		separator = "    "
	}
	return "\n" + mutedStyle.Render(trimToWidth(status, width)) + "\n" + shortcutMenu(width, separator, items)
}
