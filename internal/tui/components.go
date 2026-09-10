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
	var status string
	if m.inputMode {
		cursor := lipgloss.NewStyle().Background(colorLeaf).Foreground(colorSurface).Render(" ")
		status = "\n" + goldStyle.Render("Nueva sesion: ") + string(m.input) + cursor + "\n" + mutedStyle.Render("enter iniciar · esc cancelar")
	} else if m.err != nil {
		status = "\n" + errorStyle.Render(trimToWidth("error: "+m.err.Error(), width))
	} else if m.message != "" {
		status = "\n" + mutedStyle.Render(trimToWidth(m.message, width))
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
	content := mutedStyle.Render(title) + "\n\n" + valueStyle.Render(message)
	if m.width < compactBreakpoint || m.height < 26 {
		return "\n" + trimToWidth(title, width) + "\n" + mutedStyle.Render(trimToWidth(message, width))
	}
	return "\n" + panelStyle.Width(max(20, width-4)).Render(content)
}
