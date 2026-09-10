package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

const (
	wideBreakpoint    = 88
	compactBreakpoint = 52
)

var (
	colorLeaf    = lipgloss.Color("#A7C957")
	colorGold    = lipgloss.Color("#E9C46A")
	colorCream   = lipgloss.Color("#F4F1DE")
	colorMuted   = lipgloss.Color("#8A9185")
	colorPanel   = lipgloss.Color("#25312A")
	colorDanger  = lipgloss.Color("#E07A5F")
	colorSurface = lipgloss.Color("#18201B")

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorLeaf)
	mutedStyle = lipgloss.NewStyle().Foreground(colorMuted)
	valueStyle = lipgloss.NewStyle().Bold(true).Foreground(colorCream)
	goldStyle  = lipgloss.NewStyle().Bold(true).Foreground(colorGold)
	errorStyle = lipgloss.NewStyle().Foreground(colorDanger)
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPanel).
			Padding(0, 1)
)

type tickMsg time.Time

type Model struct {
	path      string
	db        *store.Database
	now       time.Time
	width     int
	height    int
	page      int
	help      bool
	inputMode bool
	input     []rune
	message   string
	err       error
}

func New(path string) Model {
	db, err := store.Load(path)
	if db == nil {
		db = &store.Database{}
	}
	return Model{
		path:   path,
		db:     db,
		now:    time.Now(),
		width:  80,
		height: 24,
		err:    err,
	}
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(value time.Time) tea.Msg {
		return tickMsg(value)
	})
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
	case tickMsg:
		m.now = time.Time(message)
		return m, tick()
	case tea.KeyPressMsg:
		return m.handleKey(message)
	}
	return m, nil
}

func (m Model) handleKey(message tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := message.String()
	if m.inputMode {
		switch key {
		case "esc":
			m.inputMode = false
			m.input = nil
			m.message = "cancelado"
		case "enter":
			task := strings.TrimSpace(string(m.input))
			if task == "" {
				m.message = "escribe el nombre de la tarea"
				return m, nil
			}
			if _, err := m.db.Start(time.Now(), task, ""); err != nil {
				m.err = err
			} else if err := store.Save(m.path, m.db); err != nil {
				m.err = err
			} else {
				m.message = "sesion iniciada"
				m.err = nil
				m.inputMode = false
				m.input = nil
			}
		case "backspace", "ctrl+h":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if text := message.Key().Text; text != "" {
				m.input = append(m.input, []rune(text)...)
			}
		}
		return m, nil
	}

	switch key {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "n":
		if m.db.Active() != nil {
			m.message = "termina la sesion actual primero"
			return m, nil
		}
		m.inputMode = true
		m.input = nil
		m.message = ""
		m.err = nil
	case "x":
		entry, err := m.db.Stop(time.Now(), "")
		if err != nil {
			m.err = err
		} else if err := store.Save(m.path, m.db); err != nil {
			m.err = err
		} else {
			m.message = "terminada: " + entry.Task
			m.err = nil
		}
	case "tab", "right", "l":
		m.page = (m.page + 1) % 2
	case "shift+tab", "left", "h":
		m.page = (m.page + 1) % 2
	case "?":
		m.help = !m.help
	case "r":
		db, err := store.Load(m.path)
		if err != nil {
			m.err = err
		} else {
			m.db = db
			m.err = nil
			m.message = "datos actualizados"
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	width := m.width
	if width <= 0 {
		width = 80
	}
	inner := max(10, width-2)
	if inner > 116 {
		inner = 116
	}

	header := m.header(inner)
	var body string
	if m.page == 1 {
		body = m.history(inner)
	} else {
		body = m.dashboard(inner)
	}
	footer := m.footer(inner)
	content := lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
	if m.width > inner {
		content = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(content)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func (m Model) header(width int) string {
	density := "amplia"
	if m.width < compactBreakpoint || m.height < 26 {
		density = "minima"
	} else if m.width < wideBreakpoint || m.height < 24 {
		density = "compacta"
	}
	left := titleStyle.Render("VASTAGO")
	if width >= 29 {
		left += "  " + mutedStyle.Render("crecer con intencion")
	}
	right := mutedStyle.Render(fmt.Sprintf("%s · %s", pageName(m.page), density))
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 || width < compactBreakpoint {
		return left + "\n" + mutedStyle.Render(trimToWidth(fmt.Sprintf("%s · %s", pageName(m.page), density), width))
	}
	return left + strings.Repeat(" ", gap) + right
}

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
	line := goldStyle.Render("EN PAUSA")
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
		return mutedStyle.Render("AHORA") + "\n\n" + goldStyle.Render("En pausa") + "\n" + mutedStyle.Render("Pulsa n para plantar una sesion.")
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

func (m Model) statsContent(width, taskLimit int) string {
	todayStart := store.StartOfDay(m.now)
	weekStart := todayStart.AddDate(0, 0, -6)
	entries := m.db.Since(weekStart, m.now)
	totals := store.TotalsWithin(entries, weekStart, m.now)
	result := mutedStyle.Render("RITMO") + "\n\n"
	result += statLine("Hoy", store.FormatDuration(m.totalSince(todayStart)), width) + "\n"
	result += statLine("7 dias", store.FormatDuration(m.totalSince(weekStart)), width)
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

func (m Model) footer(width int) string {
	var status string
	if m.inputMode {
		cursor := lipgloss.NewStyle().Background(colorLeaf).Foreground(colorSurface).Render(" ")
		status = "\n" + goldStyle.Render("Nueva sesion: ") + string(m.input) + cursor + "\n" + mutedStyle.Render("enter iniciar · esc cancelar")
	} else if m.err != nil {
		status = "\n" + errorStyle.Render(trimToWidth("error: "+m.err.Error(), width))
	} else if m.message != "" {
		status = "\n" + mutedStyle.Render(trimToWidth(m.message, width))
	}

	help := "n nueva · x terminar · tab vista · ? ayuda · q salir"
	if m.help {
		help = "n inicia una sesion · x termina · tab cambia vista · r recarga · q sale"
	}
	if m.width < compactBreakpoint {
		help = "n nueva · x fin · tab vista · q salir"
	}
	return status + "\n" + mutedStyle.Render(trimToWidth(help, width))
}

func (m Model) totalSince(start time.Time) time.Duration {
	var total time.Duration
	for _, entry := range m.db.Since(start, m.now) {
		total += entry.DurationWithin(start, m.now)
	}
	return total
}

func statLine(label, value string, width int) string {
	plainLabel := trimToWidth(label, max(1, width-lipgloss.Width(value)-1))
	gap := width - lipgloss.Width(plainLabel) - lipgloss.Width(value)
	if gap < 1 {
		gap = 1
	}
	return mutedStyle.Render(plainLabel) + strings.Repeat(" ", gap) + valueStyle.Render(value)
}

func pageName(page int) string {
	if page == 1 {
		return "historial"
	}
	return "inicio"
}

func trimToWidth(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	runes := []rune(value)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}
