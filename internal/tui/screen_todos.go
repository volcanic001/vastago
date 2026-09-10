package tui

func (m Model) todosScreen(width int) string {
	return m.emptyScreen(width, "PENDIENTES", "Aun no hay pendientes.")
}
