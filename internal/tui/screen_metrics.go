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
		mutedStyle.Render("RESUMEN"),
		statLine("Tiempo", store.FormatDuration(data.FocusTime), width),
		statLine("Sesiones", fmt.Sprintf("%d", data.Sessions), width),
		statLine("Pendientes hechos", fmt.Sprintf("%d", data.CompletedTodos), width),
	}
	if data.HabitOpportunities == 0 {
		lines = append(lines, statLine("Habitos", "sin datos", width))
	} else {
		lines = append(lines, statLine("Habitos", fmt.Sprintf("%d/%d · %.0f%%", data.HabitCompletions, data.HabitOpportunities, data.HabitPercent()), width))
	}
	previousPeriod, err := p.Shift(-1)
	if err != nil {
		return errorStyle.Render(trimToWidth(err.Error(), width)), 0
	}
	currentActivity, err := m.db.Activity(p, m.now)
	if err != nil {
		return errorStyle.Render(trimToWidth(err.Error(), width)), 0
	}
	previousActivity, err := m.db.Activity(previousPeriod, m.now)
	if err != nil {
		return errorStyle.Render(trimToWidth(err.Error(), width)), 0
	}
	if chart := comparisonChart(width, buildComparisonSeries(p, currentActivity, previousActivity)); len(chart) > 0 {
		lines = append(lines, "")
		lines = append(lines, chart...)
	}
	if data.FocusTime == 0 {
		message := "Sin actividad registrada en este periodo."
		if p.Start.After(m.now) {
			message = "Periodo futuro: sin actividad."
		}
		lines = append(lines, "", mutedStyle.Render(message))
	} else {
		lines = append(lines, "", mutedStyle.Render("TIEMPO POR TAREA"))
		for _, task := range data.Tasks {
			fraction := float64(task.Duration) / float64(data.FocusTime)
			lines = append(lines, task.Task+" · "+store.FormatDuration(task.Duration)+" · "+progressPercent(fraction))
			lines = append(lines, focusProgressBar(width, fraction))
		}
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
	hints := "d/w/m/y periodo · [/] mover · t actual · j/k o ↑↓ · tab vistas · q salir"
	if width < 56 {
		hints = "d/w/m/y · [/] · t · j/k o ↑↓ · tab · q"
	}
	if width < 34 {
		hints = "d/w/m/y · [/] · t\nj/k o ↑↓ · tab · q"
	}
	if m.help {
		hints += "\nHabitos: incluye hoy. Comparativa: actual frente al periodo anterior."
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
