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
		items := []shortcut{{key: "enter", action: "iniciar", primary: true}, {key: "esc", action: "cancelar"}}
		if m.inputAction == inputTodoNew {
			label = "Nuevo pendiente: "
		} else if m.inputAction == inputTodoEdit {
			label = "Editar pendiente: "
			items = []shortcut{{key: "enter", action: "guardar", primary: true}, {key: "esc", action: "cancelar"}, {key: "ctrl+u", action: "limpiar"}}
		} else if m.inputAction == inputHabitNew {
			label = "Nuevo habito: "
			items = []shortcut{{key: "enter", action: "guardar", primary: true}, {key: "esc", action: "cancelar"}, {key: "ctrl+u", action: "limpiar"}}
		} else if m.inputAction == inputHabitEdit {
			label = "Editar habito: "
			items = []shortcut{{key: "enter", action: "guardar", primary: true}, {key: "esc", action: "cancelar"}, {key: "ctrl+u", action: "limpiar"}}
		}
		cursor := lipgloss.NewStyle().Background(colorLeaf).Foreground(colorSurface).Render(" ")
		status = "\n" + goldStyle.Render(label) + string(m.input) + cursor
		return status + "\n" + shortcutMenu(width, " · ", items)
	}
	if m.err != nil {
		status = "\n" + errorStyle.Render(trimToWidth("error: "+m.err.Error(), width))
	} else if m.message != "" {
		status = "\n" + mutedStyle.Render(trimToWidth(m.message, width))
	}

	if m.confirmHabit || m.confirmTodo {
		return status + "\n" + shortcutMenu(width, "    ", []shortcut{{key: "y", action: "borrar", tone: shortcutDanger, primary: true}, {key: "n/esc", action: "cancelar"}})
	}
	if m.screen == habitsScreen {
		return status + "\n" + shortcutMenu(width, " · ", []shortcut{{key: "j/k/↑↓", action: "mover"}, {key: "space", action: "marcar hoy"}, {key: "n", action: "nuevo", primary: true}, {key: "e", action: "editar", primary: true}, {key: "d", action: "borrar", tone: shortcutDanger, primary: true}, {key: "tab", action: "vistas"}, {key: "q", action: "salir", primary: true}})
	}
	if m.screen == todosScreen {
		return status + "\n" + shortcutMenu(width, " · ", []shortcut{{key: "j/k/↑↓", action: "mover"}, {key: "space", action: "completar/reabrir"}, {key: "n", action: "nuevo", primary: true}, {key: "e", action: "editar", primary: true}, {key: "d", action: "borrar", tone: shortcutDanger, primary: true}, {key: "tab", action: "vistas"}, {key: "q", action: "salir", primary: true}})
	}
	if m.screen == homeScreen {
		return status + "\n" + shortcutMenu(width, " · ", []shortcut{{key: "n", action: "iniciar", primary: true}, {key: "x", action: "fin", primary: true}, {key: "2", action: "pendientes"}, {key: "3", action: "habitos"}, {key: "4", action: "sesiones"}, {key: "5", action: "metricas"}, {key: "q", action: "salir", primary: true}})
	}
	items := []shortcut{{key: "1-5", action: "vistas"}, {key: "n", action: "nueva"}, {key: "x", action: "terminar"}, {key: "?", action: "ayuda"}, {key: "q", action: "salir", primary: true}}
	footer := status + "\n" + shortcutMenu(width, " · ", items)
	if m.help {
		footer += "\n" + mutedStyle.Render(trimToWidth("1-5 o tab cambian vista · n inicia · x termina · r recarga · q sale", width))
	}
	return footer
}

type shortcutTone int

const (
	shortcutNormal shortcutTone = iota
	shortcutDanger
)

type shortcut struct {
	key     string
	action  string
	tone    shortcutTone
	primary bool
}

func (s shortcut) label() string {
	return "[" + s.key + "] " + s.action
}

func (s shortcut) render() string {
	key := "[" + s.key + "]"
	if s.tone == shortcutDanger {
		key = errorStyle.Render(key)
	} else {
		key = shortcutKeyStyle.Render(key)
	}
	return key + mutedStyle.Render(" "+s.action)
}

func shortcutMenu(width int, separator string, items []shortcut) string {
	if len(items) == 0 {
		return ""
	}
	if width <= 0 || shortcutMenuWidth(items, separator) <= width {
		return renderShortcutItems(items, separator)
	}

	primary, secondary := prioritizedShortcutRows(items)
	if len(primary) > 0 && len(secondary) > 0 && shortcutMenuWidth(primary, separator)*2 < width {
		return renderShortcutItems(primary, separator) + "\n" + renderShortcutItems(secondary, separator)
	}

	split := compactShortcutSplit(width, separator, items)
	return renderShortcutItems(items[:split], separator) + "\n" + renderShortcutItems(items[split:], separator)
}

func prioritizedShortcutRows(items []shortcut) ([]shortcut, []shortcut) {
	primary := make([]shortcut, 0, len(items))
	secondary := make([]shortcut, 0, len(items))
	for _, item := range items {
		if item.primary {
			primary = append(primary, item)
		} else {
			secondary = append(secondary, item)
		}
	}
	if len(primary) == 0 || len(secondary) == 0 {
		return nil, nil
	}
	return primary, secondary
}

func compactShortcutSplit(width int, separator string, items []shortcut) int {
	bestSplit, bestOverflow := 1, -1
	for candidate := 1; candidate < len(items); candidate++ {
		left := shortcutMenuWidth(items[:candidate], separator)
		right := shortcutMenuWidth(items[candidate:], separator)
		if left <= width && right <= width {
			return candidate
		}
		overflow := max(0, left-width) + max(0, right-width)
		if bestOverflow < 0 || overflow < bestOverflow {
			bestSplit, bestOverflow = candidate, overflow
		}
	}
	return bestSplit
}

func shortcutMenuWidth(items []shortcut, separator string) int {
	width := 0
	for index, item := range items {
		if index > 0 {
			width += lipgloss.Width(separator)
		}
		width += lipgloss.Width(item.label())
	}
	return width
}

func renderShortcutItems(items []shortcut, separator string) string {
	rendered := make([]string, len(items))
	for index, item := range items {
		rendered[index] = item.render()
	}
	return strings.Join(rendered, mutedStyle.Render(separator))
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
