package tui

type screen int

const (
	sessionsScreen screen = iota
	todosScreen
	habitsScreen
	metricsScreen
	screenCount
)

func (m Model) nextScreen() Model {
	m.screen = (m.screen + 1) % screenCount
	return m.clearFeedback()
}

func (m Model) previousScreen() Model {
	m.screen = (m.screen - 1 + screenCount) % screenCount
	return m.clearFeedback()
}

func (m Model) selectScreen(selected screen) Model {
	if selected >= sessionsScreen && selected < screenCount && selected != m.screen {
		m.screen = selected
		return m.clearFeedback()
	}
	return m
}

func (m Model) clearFeedback() Model {
	m.message, m.err = "", nil
	return m
}

func (s screen) name() string {
	switch s {
	case sessionsScreen:
		return "sesiones"
	case todosScreen:
		return "pendientes"
	case habitsScreen:
		return "habitos"
	case metricsScreen:
		return "metricas"
	default:
		return "sesiones"
	}
}
