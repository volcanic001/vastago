package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/volcanic001/vastago/internal/store"
)

type inputAction int

const (
	inputSession inputAction = iota
	inputTodoNew
	inputTodoEdit
	inputHabitNew
	inputHabitEdit
)

func (m Model) handleKey(message tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.screen == metricsScreen && !m.inputMode {
		if updated, handled := m.handleMetricsKey(message); handled {
			return updated, nil
		}
	}
	if !m.inputMode && !m.confirmTodo && !m.confirmHabit && m.screen == sessionsScreen {
		if updated, handled := m.handleSessionKey(message); handled {
			return updated, nil
		}
	}
	if m.inputMode {
		return m.handleInput(message), nil
	}
	if m.confirmTodo {
		return m.handleTodoConfirmation(message), nil
	}

	if m.confirmHabit {
		return m.handleHabitConfirmation(message), nil
	}

	switch message.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "n":
		if m.screen == todosScreen {
			m.inputMode, m.inputAction, m.input = true, inputTodoNew, nil
			m.message, m.err = "", nil
			return m, nil
		}
		if m.screen == habitsScreen {
			m.inputMode, m.inputAction, m.input = true, inputHabitNew, nil
			m.message, m.err = "", nil
			return m, nil
		}
		if m.db.Active() != nil {
			m.message = "termina la sesion actual primero"
			return m, nil
		}
		m.inputMode, m.inputAction, m.input = true, inputSession, nil
		m.message, m.err = "", nil
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
	case "j", "down":
		if m.screen == todosScreen && m.selectedTodo < len(m.db.Todos)-1 {
			m.selectedTodo++
		} else if m.screen == habitsScreen && m.selectedHabit < len(m.db.Habits)-1 {
			m.selectedHabit++
		}
	case "k", "up":
		if m.screen == todosScreen && m.selectedTodo > 0 {
			m.selectedTodo--
		} else if m.screen == habitsScreen && m.selectedHabit > 0 {
			m.selectedHabit--
		}
	case "space":
		if m.screen == todosScreen {
			m = m.toggleSelectedTodo()
		} else if m.screen == habitsScreen {
			m = m.toggleSelectedHabit()
		}
	case "e":
		if m.screen == todosScreen && len(m.db.Todos) > 0 {
			m.inputMode, m.inputAction = true, inputTodoEdit
			m.input = []rune(m.db.Todos[m.selectedTodo].Title)
			m.message, m.err = "", nil
		} else if m.screen == habitsScreen && len(m.db.Habits) > 0 {
			m.inputMode, m.inputAction = true, inputHabitEdit
			m.input = []rune(m.db.Habits[m.selectedHabit].Name)
			m.message, m.err = "", nil
		}
	case "d":
		if m.screen == todosScreen && len(m.db.Todos) > 0 {
			m.confirmTodo = true
			m.message = "eliminar pendiente? y confirmar · n cancelar"
			m.err = nil
		} else if m.screen == habitsScreen && len(m.db.Habits) > 0 {
			m.confirmHabit = true
			m.message = "eliminar habito? y confirmar · n cancelar"
			m.err = nil
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
			m.clampTodoSelection()
			m.clampHabitSelection()
		}
	}
	return m, nil
}

func (m Model) handleInput(message tea.KeyPressMsg) Model {

	switch message.String() {
	case "esc":
		m.inputMode, m.input, m.message = false, nil, "cancelado"
	case "enter":
		value := strings.TrimSpace(string(m.input))
		if value == "" {
			m.message = "escribe un nombre"
			return m
		}
		if m.inputAction == inputHabitNew {
			return m.createHabit(value)
		}
		if m.inputAction == inputHabitEdit {
			return m.renameSelectedHabit(value)
		}
		switch m.inputAction {
		case inputTodoNew:
			if _, err := m.db.AddTodo(time.Now(), value); err != nil {
				m.err = err
			} else if err := store.Save(m.path, m.db); err != nil {
				m.err = err
			} else {
				m.selectedTodo = len(m.db.Todos) - 1
				m.finishInput("pendiente creado")
			}
		case inputTodoEdit:
			if len(m.db.Todos) == 0 {
				m.finishInput("sin pendientes")
				return m
			}
			id := m.db.Todos[m.selectedTodo].ID
			if _, err := m.db.RenameTodo(id, value); err != nil {
				m.err = err
			} else if err := store.Save(m.path, m.db); err != nil {
				m.err = err
			} else {
				m.finishInput("pendiente actualizado")
			}
		default:
			if _, err := m.db.Start(time.Now(), value, ""); err != nil {
				m.err = err
			} else if err := store.Save(m.path, m.db); err != nil {
				m.err = err
			} else {
				m.finishInput("sesion iniciada")
			}
		}
	case "backspace", "ctrl+h":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case "ctrl+u":
		m.input = nil
	default:
		if text := message.Key().Text; text != "" {
			m.input = append(m.input, []rune(text)...)
		}
	}
	return m
}

func (m *Model) finishInput(message string) {
	m.inputMode, m.input, m.message, m.err = false, nil, message, nil
}

func (m Model) toggleSelectedTodo() Model {
	if len(m.db.Todos) == 0 {
		return m
	}
	todo, err := m.db.ToggleTodo(time.Now(), m.db.Todos[m.selectedTodo].ID)
	if err != nil {
		m.err = err
	} else if err := store.Save(m.path, m.db); err != nil {
		m.err = err
	} else if todo.Completed() {
		m.message, m.err = "pendiente completado", nil
	} else {
		m.message, m.err = "pendiente reabierto", nil
	}
	return m
}

func (m Model) handleTodoConfirmation(message tea.KeyPressMsg) Model {

	switch message.String() {
	case "y", "Y":
		if len(m.db.Todos) > 0 {
			err := m.db.DeleteTodo(m.db.Todos[m.selectedTodo].ID)
			if err == nil {
				err = store.Save(m.path, m.db)
			}
			if err != nil {
				m.err = err
			} else {
				m.message, m.err = "pendiente eliminado", nil
				m.clampTodoSelection()
				m.clampHabitSelection()
			}
		}
		m.confirmTodo = false
	case "n", "N", "esc":
		m.confirmTodo = false
		m.message = "eliminacion cancelada"
	}
	return m
}

func (m *Model) clampTodoSelection() {
	if len(m.db.Todos) == 0 {
		m.selectedTodo = 0
	} else if m.selectedTodo >= len(m.db.Todos) {
		m.selectedTodo = len(m.db.Todos) - 1
	}
}
