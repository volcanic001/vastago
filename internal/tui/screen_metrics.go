package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func (m Model) selectedPeriod() (store.Period, error) {
	kind := m.metricsKind
	if kind == "" {
		kind = store.Week
	}
	anchor := m.metricsAnchor
	if anchor.IsZero() {
		anchor = m.now
	}
	return store.NewPeriod(kind, anchor.In(time.Local))
}

func (m Model) handleMetricsKey(key tea.KeyPressMsg) (Model, bool) {
	switch key.String() {
	case "d", "w", "m", "y":
		m.metricsKind = map[string]store.PeriodKind{"d": store.Day, "w": store.Week, "m": store.Month, "y": store.Year}[key.String()]
	case "[", "]":
		p, err := m.selectedPeriod()
		if err != nil {
			m.err = err
			return m, true
		}
		offset := 1
		if key.String() == "[" {
			offset = -1
		}
		p, err = p.Shift(offset)
		if err != nil {
			m.err = err
			return m, true
		}
		m.metricsAnchor = p.Start
	case "t":
		m.metricsAnchor = time.Time{}
	case "j", "down":
		_, limit := m.metricsPage(max(1, m.width))
		m.metricsScroll = min(m.metricsScroll+1, limit)
		return m, true
	case "k", "up":
		m.metricsScroll = max(0, m.metricsScroll-1)
		return m, true
	default:
		return m, false
	}
	m.metricsScroll = 0
	m.err, m.message = nil, ""
	return m, true
}

func periodLabel(p store.Period) string {
	switch p.Kind {
	case store.Day:
		return "DIA · " + p.Start.Format("02/01/2006")
	case store.Week:
		return "SEMANA · " + p.Start.Format("02/01/2006") + " - " + p.End.AddDate(0, 0, -1).Format("02/01/2006")
	case store.Month:
		return "MES · " + p.Start.Format("01/2006")
	default:
		return "AÑO · " + p.Start.Format("2006")
	}
}

func (m Model) metricsView(width int) string {
	content, _ := m.metricsPage(width)
	return content
}

func (m Model) metricsPage(width int) (string, int) {
	p, err := m.selectedPeriod()
	if err != nil {
		return errorStyle.Render(trimToWidth(err.Error(), width)), 0
	}
	data, err := m.db.Metrics(p, m.now)
	if err != nil {
		return errorStyle.Render(trimToWidth(err.Error(), width)), 0
	}
	lines := []string{
		"Tiempo: " + store.FormatDuration(data.FocusTime),
		fmt.Sprintf("Sesiones: %d", data.Sessions),
		fmt.Sprintf("Pendientes hechos: %d", data.CompletedTodos),
	}
	if data.HabitOpportunities == 0 {
		lines = append(lines, "Habitos: sin datos")
	} else {
		lines = append(lines, fmt.Sprintf("Habitos: %d/%d · %.0f%%", data.HabitCompletions, data.HabitOpportunities, data.HabitPercent()))
	}
	if p.Start.After(m.now) {
		lines = append(lines, "Periodo futuro: sin actividad")
	}
	lines = append(lines, "", "TIEMPO POR TAREA")
	if len(data.Tasks) == 0 {
		lines = append(lines, "Sin tiempo registrado.")
	}
	for _, task := range data.Tasks {
		lines = append(lines, task.Task+" · "+store.FormatDuration(task.Duration))
		fraction := float64(task.Duration) / float64(data.FocusTime)
		cells := max(1, min(width, int(fraction*float64(width))))
		lines = append(lines, titleStyle.Render(strings.Repeat("█", cells)))
	}
	// Wrap before scrolling so long labels remain readable on narrow terminals.
	wrapped := strings.Split(ansi.Wrap(strings.Join(lines, "\n"), max(1, width), ""), "\n")
	title := strings.Split(ansi.Wrap(periodLabel(p), max(1, width), ""), "\n")
	height := m.height
	if height <= 0 {
		height = 24
	}
	available := max(1, height-lipgloss.Height(m.header(width))-lipgloss.Height(m.metricsFooter(width))-1-len(title))
	offset := min(m.metricsScroll, max(0, len(wrapped)-available))
	visible := wrapped[offset:min(len(wrapped), offset+available)]
	return "\n" + strings.Join(title, "\n") + "\n" + strings.Join(visible, "\n"), max(0, len(wrapped)-available)
}

func (m Model) metricsFooter(width int) string {
	hints := "d dia · w semana · m mes · y año\n[ anterior · ] siguiente · t actual\nj/k desplazar · tab vistas · q salir"
	if width < 40 {
		hints = "d/w/m/y periodo\n[/] cambiar · t hoy\nj/k mover · tab · q"
	}
	if m.help {
		hints += "\nHabitos: incluye hoy. Barras: parte del tiempo total."
	}
	if m.err != nil {
		hints = "error: " + m.err.Error() + "\n" + hints
	}
	rows := strings.Split(hints, "\n")
	for i := range rows {
		rows[i] = trimToWidth(rows[i], width)
	}
	return "\n" + mutedStyle.Render(strings.Join(rows, "\n"))
}
