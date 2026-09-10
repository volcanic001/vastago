package tui

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"strings"
)

func (m Model) header(width int) string {
	left := titleStyle.Render("VASTAGO")
	if width >= 29 {
		left += "  " + mutedStyle.Render("crecer con intencion")
	}
	right := mutedStyle.Render(fmt.Sprintf("%s · %s", m.now.Format("02/01 15:04"), m.screen.name()))
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 || width < compactBreakpoint {
		return left + "\n" + mutedStyle.Render(trimToWidth(fmt.Sprintf("%s · %s", m.now.Format("02/01 15:04"), m.screen.name()), width))
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) footer(width int) string {
	if m.screen == metricsScreen && !m.inputMode {
		return m.metricsFooter(width)
	}
	if m.screen == sessionsScreen && !m.inputMode {
		return m.sessionFooter(width)
	}
	var status string
	if m.inputMode {
		label := "Nueva sesion: "
		if m.inputAction == inputTodoNew {
			label = "Nuevo pendiente: "
		} else if m.inputAction == inputTodoEdit {
			label = "Editar pendiente: "
		} else if m.inputAction == inputHabitNew {
			label = "Nuevo habito: "
		} else if m.inputAction == inputHabitEdit {
			label = "Editar habito: "
		}
		hint := "enter iniciar · esc cancelar"
		if m.inputAction != inputSession {
			hint = "enter guardar · esc cancelar · ctrl+u limpiar"
		}
		cursor := lipgloss.NewStyle().Background(colorLeaf).Foreground(colorSurface).Render(" ")
		status = "\n" + goldStyle.Render(label) + string(m.input) + cursor + "\n" + mutedStyle.Render(hint)
	} else if m.err != nil {
		status = "\n" + errorStyle.Render(trimToWidth("error: "+m.err.Error(), width))
	} else if m.message != "" {
		status = "\n" + mutedStyle.Render(trimToWidth(m.message, width))
	}

	if m.screen == habitsScreen && !m.inputMode {
		help := "↑↓ mover · space marcar hoy · n nuevo · e editar · d borrar · q salir"
		if m.width < compactBreakpoint {
			help = "↑↓ mover · space hoy · n nuevo · e editar · d borrar"
		}
		return status + "\n" + mutedStyle.Render(trimToWidth(help, width))
	}

	if m.screen == todosScreen && !m.inputMode {
		help := "↑↓ mover · space completar · n nuevo · e editar · d borrar · q salir"
		if m.width < compactBreakpoint {
			help = "↑↓ mover · space hecho · n nuevo · e editar · d borrar"
		}
		return status + "\n" + mutedStyle.Render(trimToWidth(help, width))
	}

	help := "1-5 vistas · n nueva · x terminar · ? ayuda · q salir"
	if m.help {
		help = "1-5 o tab cambian vista · n inicia · x termina · r recarga · q sale"
	}
	if m.width < compactBreakpoint {
		help = "1-5 vista · n nueva · x fin · q salir"
	}
	return status + "\n" + mutedStyle.Render(trimToWidth(help, width))
}

func statLine(label, value string, width int) string {
	plainLabel := trimToWidth(label, max(1, width-lipgloss.Width(value)-1))
	gap := width - lipgloss.Width(plainLabel) - lipgloss.Width(value)
	if gap < 1 {
		gap = 1
	}
	return mutedStyle.Render(plainLabel) + strings.Repeat(" ", gap) + valueStyle.Render(value)
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

func (m Model) emptyScreen(width int, title, message string) string {
	return "\n" + mutedStyle.Render(trimToWidth(title, width)) + "\n" + valueStyle.Render(trimToWidth(message, width))
}
