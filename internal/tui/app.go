package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

type tickMsg time.Time

type Model struct {
	selectedSession int
	sessionEdit     *sessionForm
	sessionDeleteID string
	path            string
	db              *store.Database
	now             time.Time
	width           int
	height          int
	screen          screen
	selectedTodo    int
	selectedHabit   int
	inputAction     inputAction
	confirmTodo     bool
	confirmHabit    bool
	help            bool
	inputMode       bool
	input           []rune
	message         string
	err             error
}

func New(path string) Model {
	db, err := store.Load(path)
	if db == nil {
		db = &store.Database{}
	}
	return Model{path: path, db: db, now: time.Now(), width: 80, height: 24, err: err}
}

func (m Model) Init() tea.Cmd { return tick() }

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(value time.Time) tea.Msg { return tickMsg(value) })
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
	case tickMsg:
		m.now = time.Time(message)
		return m, tick()
	case tea.KeyPressMsg:
		return m.handleKey(message)
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

	var body string
	switch m.screen {
	case sessionsScreen:
		body = m.sessionView(inner)
	case todosScreen:
		body = m.todosScreen(inner)
	case habitsScreen:
		body = m.habitsScreen(inner)
	case metricsScreen:
		body = m.metricsScreen(inner)
	default:
		body = m.dashboard(inner)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, m.header(inner), body, m.footer(inner))
	if m.width > inner {
		content = lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Render(content)
	}
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}
