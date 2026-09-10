package tui

func (m Model) habitsScreen(width int) string {
	return m.emptyScreen(width, "HABITOS", "Aun no hay habitos.")
}
