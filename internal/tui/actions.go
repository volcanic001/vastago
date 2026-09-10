package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/volcanic001/vastago/internal/store"
)

func (m Model) handleKey(message tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.inputMode {
		return m.handleInput(message), nil
	}
	switch message.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "n":
		if m.db.Active() != nil {
			m.message = "termina la sesion actual primero"
			return m, nil
		}
		m.inputMode, m.input, m.message, m.err = true, nil, "", nil
	case "x":
		if m.db.Active() == nil {
			m.err, m.message = nil, "Sin sesión activa."
			return m, nil
		}
		entry, err := m.db.Stop(time.Now(), "")
		if err != nil {
			m.err = err
		} else if err := store.Save(m.path, m.db); err != nil {
			m.err = err
		} else {
			m.message, m.err = "terminada: "+entry.Task, nil
		}
	case "tab", "right", "l":
		m = m.nextScreen()
	case "shift+tab", "left", "h":
		m = m.previousScreen()
	case "1":
		m = m.selectScreen(homeScreen)
	case "2":
		m = m.selectScreen(sessionsScreen)
	case "3":
		m = m.selectScreen(todosScreen)
	case "4":
		m = m.selectScreen(habitsScreen)
	case "5":
		m = m.selectScreen(metricsScreen)
	case "?":
		m.help = !m.help
	case "r":
		db, err := store.Load(m.path)
		if err != nil {
			m.err = err
		} else {
			m.db, m.err, m.message = db, nil, "datos actualizados"
		}
	}
	return m, nil
}

func (m Model) handleInput(message tea.KeyPressMsg) Model {
	switch message.String() {
	case "esc":
		m.inputMode, m.input, m.message = false, nil, "cancelado"
	case "enter":
		task := strings.TrimSpace(string(m.input))
		if task == "" {
			m.message = "escribe el nombre de la tarea"
			return m
		}
		if _, err := m.db.Start(time.Now(), task, ""); err != nil {
			m.err = err
		} else if err := store.Save(m.path, m.db); err != nil {
			m.err = err
		} else {
			m.message, m.err, m.inputMode, m.input = "sesion iniciada", nil, false, nil
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
	return m
}
