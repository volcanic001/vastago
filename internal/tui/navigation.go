package tui

type screen int

const (
	homeScreen screen = iota
	sessionsScreen
)

func (m Model) nextScreen() Model {
	m.screen = (m.screen + 1) % 2
	return m
}

func (m Model) previousScreen() Model {
	m.screen = (m.screen + 1) % 2
	return m
}

func (s screen) name() string {
	if s == sessionsScreen {
		return "historial"
	}
	return "inicio"
}
