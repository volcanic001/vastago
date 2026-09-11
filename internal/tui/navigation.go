package tui

type screen int

const (
	homeScreen screen = iota
	todosScreen
	habitsScreen
	sessionsScreen
	metricsScreen
	screenCount
)

func (m Model) nextScreen() Model {
	m.screen = (m.screen + 1) % screenCount
	return m
}

func (m Model) previousScreen() Model {
	m.screen = (m.screen - 1 + screenCount) % screenCount
	return m
}

func (m Model) selectScreen(selected screen) Model {
	if selected >= homeScreen && selected < screenCount {
		m.screen = selected
	}
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
		return "inicio"
	}
}
