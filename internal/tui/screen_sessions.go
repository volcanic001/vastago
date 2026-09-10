package tui

import (
	"fmt"
	"github.com/volcanic001/vastago/internal/store"
	"strings"
)

func (m Model) sessionView(width int) string {
	if m.sessionEdit != nil {
		return m.sessionFormView(width)
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
		state := store.FormatDuration(entry.Duration(m.now))
		if entry.End == nil {
			state += " activa"
		}
		line := trimToWidth(fmt.Sprintf("%s %s · %s", entry.Start.Format("02/01"), entry.Task, state), width)
		if row == selected {
			line = selectionStyle.Width(width).Render(line)
		}
		lines = append(lines, line)
	}
	return "\n" + strings.Join(lines, "\n")
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
	hint := "j/k o ↑↓ mover · enter/e editar · d borrar · n nueva · x fin · tab vistas · q salir"
	if m.sessionEdit != nil {
		hint = "enter siguiente/guardar · tab campo · ctrl+u limpiar · esc cancelar"
	}
	if m.sessionDeleteID != "" {
		hint = "y borrar · n/esc cancelar"
	}
	status := m.message
	if m.err != nil {
		status = "error: " + m.err.Error()
	}
	return "\n" + mutedStyle.Render(trimToWidth(status, width)) + "\n" + mutedStyle.Render(trimToWidth(hint, width))
}
