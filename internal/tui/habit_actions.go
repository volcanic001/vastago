package tui

import (
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/volcanic001/vastago/internal/store"
)

func (m Model) createHabit(name string) Model {
	if _, err := m.db.AddHabit(time.Now(), name); err != nil {
		m.err = err
	} else if err := store.Save(m.path, m.db); err != nil {
		m.err = err
	} else {
		m.selectedHabit = len(m.db.Habits) - 1
		m.finishInput("habito creado")
	}
	return m
}

func (m Model) renameSelectedHabit(name string) Model {
	if len(m.db.Habits) == 0 {
		m.finishInput("sin habitos")
		return m
	}
	id := m.db.Habits[m.selectedHabit].ID
	if _, err := m.db.RenameHabit(id, name); err != nil {
		m.err = err
	} else if err := store.Save(m.path, m.db); err != nil {
		m.err = err
	} else {
		m.finishInput("habito actualizado")
	}
	return m
}

func (m Model) toggleSelectedHabit() Model {
	if len(m.db.Habits) == 0 {
		return m
	}
	habit, err := m.db.ToggleHabit(m.now, m.db.Habits[m.selectedHabit].ID)
	if err != nil {
		m.err = err
	} else if err := store.Save(m.path, m.db); err != nil {
		m.err = err
	} else if habit.CompletedOn(m.now) {
		m.message, m.err = "habito marcado por hoy", nil
	} else {
		m.message, m.err = "marca de hoy eliminada", nil
	}
	return m
}

func (m Model) handleHabitConfirmation(message tea.KeyPressMsg) Model {
	switch message.String() {
	case "y", "Y":
		if len(m.db.Habits) > 0 {
			err := m.db.DeleteHabit(m.db.Habits[m.selectedHabit].ID)
			if err == nil {
				err = store.Save(m.path, m.db)
			}
			if err != nil {
				m.err = err
			} else {
				m.message, m.err = "habito eliminado", nil
				m.clampHabitSelection()
			}
		}
		m.confirmHabit = false
	case "n", "N", "esc":
		m.confirmHabit = false
		m.message = "eliminacion cancelada"
	}
	return m
}

func (m *Model) clampHabitSelection() {
	if len(m.db.Habits) == 0 {
		m.selectedHabit = 0
	} else if m.selectedHabit >= len(m.db.Habits) {
		m.selectedHabit = len(m.db.Habits) - 1
	}
}
