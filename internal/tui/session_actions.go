package tui

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/volcanic001/vastago/internal/store"
	"time"
)

const sessionDateFormat = "2006-01-02 15:04:05"

type sessionForm struct {
	id     string
	values [4]string
	field  int
	active bool
	start  time.Time
	end    *time.Time
}

func (m Model) handleSessionKey(key tea.KeyPressMsg) (Model, bool) {
	m.selectedSession = max(0, min(m.selectedSession, len(m.db.Entries)-1))
	if m.sessionEdit != nil {
		return m.editSessionKey(key), true
	}
	if m.sessionRepeatID != "" {
		return m.handleSessionRepeatConfirmation(key), true
	}
	if m.sessionDetailID != "" {
		switch key.String() {
		case "i", "esc", "enter":
			m.sessionDetailID = ""
		}
		return m, true
	}
	if m.sessionDeleteID != "" {
		switch key.String() {
		case "y", "Y":
			db := m.sessionCopy()
			err := db.DeleteSession(m.sessionDeleteID)
			if err == nil {
				err = store.Save(m.path, db)
			}
			if err != nil {
				m.err = err
				return m, true
			}
			m.db = db
			m.selectedSession = max(0, min(m.selectedSession, len(db.Entries)-1))
			m.sessionDeleteID = ""
			m.message, m.err = "sesion eliminada", nil
		case "n", "N", "esc":
			m.sessionDeleteID = ""
			m.message, m.err = "eliminacion cancelada", nil
		}
		return m, true
	}
	switch key.String() {
	case "j", "down":
		m.selectedSession = min(m.selectedSession+1, max(0, len(m.db.Entries)-1))
	case "k", "up":
		m.selectedSession = max(0, m.selectedSession-1)
	case "r":
		if len(m.db.Entries) == 0 {
			return m, true
		}
		entry := m.db.Entries[len(m.db.Entries)-1-m.selectedSession]
		if active := m.db.Active(); active != nil {
			if active.ID == entry.ID {
				m.message, m.err = "esa sesion ya esta activa", nil
				return m, true
			}
			m.sessionRepeatID = entry.ID
			m.message, m.err = "", nil
			return m, true
		}
		return m.startRepeatedSession(entry.Task), true
	case "e", "enter", "d", "i":
		if len(m.db.Entries) == 0 {
			return m, true
		}
		entry := m.db.Entries[len(m.db.Entries)-1-m.selectedSession]
		m.err, m.message = nil, ""

		if key.String() == "i" {
			m.sessionDetailID = entry.ID
			return m, true
		}
		if key.String() == "d" {
			m.sessionDeleteID = entry.ID
			return m, true
		}
		form := &sessionForm{id: entry.ID, active: entry.End == nil, start: entry.Start, end: entry.End}
		form.values = [4]string{entry.Task, entry.Note, entry.Start.Local().Format(sessionDateFormat), ""}
		if entry.End != nil {
			form.values[3] = entry.End.Local().Format(sessionDateFormat)
		}
		m.sessionEdit = form
	default:
		return m, false
	}
	return m, true
}

func (m Model) startRepeatedSession(task string) Model {
	db := m.sessionCopy()
	if _, err := db.Start(m.now, task, ""); err != nil {
		m.err = err
		return m
	}
	if err := store.Save(m.path, db); err != nil {
		m.err = err
		return m
	}
	m.db = db
	m.selectedSession = 0
	m.sessionRepeatID = ""
	m.message, m.err = "sesion repetida: "+task, nil
	return m
}

func (m Model) handleSessionRepeatConfirmation(key tea.KeyPressMsg) Model {
	switch key.String() {
	case "y", "Y":
		target, ok := m.sessionByID(m.sessionRepeatID)
		if !ok {
			m.sessionRepeatID = ""
			m.err = store.ErrSessionNotFound
			return m
		}
		db := m.sessionCopy()
		if _, err := db.Stop(m.now, ""); err != nil {
			m.err = err
			return m
		}
		if _, err := db.Start(m.now, target.Task, ""); err != nil {
			m.err = err
			return m
		}
		if err := store.Save(m.path, db); err != nil {
			m.err = err
			return m
		}
		m.db = db
		m.selectedSession = 0
		m.sessionRepeatID = ""
		m.message, m.err = "sesion repetida: "+target.Task, nil
	case "n", "N", "esc":
		m.sessionRepeatID = ""
		m.message, m.err = "cambio cancelado", nil
	}
	return m
}

func (m Model) sessionByID(id string) (store.Entry, bool) {
	for _, entry := range m.db.Entries {
		if entry.ID == id {
			return entry, true
		}
	}
	return store.Entry{}, false
}

func (m Model) sessionCopy() *store.Database {
	db := *m.db
	db.Entries = append([]store.Entry(nil), m.db.Entries...)
	return &db
}

func (m Model) editSessionKey(key tea.KeyPressMsg) Model {
	form := *m.sessionEdit
	m.sessionEdit = &form
	count := 4
	if form.active {
		count = 3
	}
	switch key.String() {
	case "esc":
		m.sessionEdit = nil
		m.err, m.message = nil, "edicion cancelada"
	case "up", "shift+tab":
		form.field = (form.field + count - 1) % count
	case "down", "tab":
		form.field = (form.field + 1) % count
	case "ctrl+s":
		return m.saveSessionEdit(&form)
	case "enter":
		if form.field < count-1 {
			form.field++
			return m
		}
		return m.saveSessionEdit(&form)
	case "ctrl+u":
		form.values[form.field] = ""
	case "backspace", "ctrl+h":
		runes := []rune(form.values[form.field])
		if len(runes) > 0 {
			form.values[form.field] = string(runes[:len(runes)-1])
		}
	default:
		form.values[form.field] += key.Key().Text
	}
	return m
}

func (m Model) saveSessionEdit(form *sessionForm) Model {
	start, err := parseSessionDate(form.values[2], form.start)
	if err != nil {
		m.err = err
		form.field = 2
		return m
	}
	var end *time.Time
	if !form.active {
		value, err := parseSessionDate(form.values[3], *form.end)
		if err != nil {
			m.err = err
			form.field = 3
			return m
		}
		end = &value
	}
	if form.active && start.After(m.now) {
		m.err = fmt.Errorf("el inicio activo no puede estar en el futuro")
		return m
	}
	db := m.sessionCopy()
	err = db.EditSession(form.id, form.values[0], form.values[1], start, end, m.now)
	if err == nil {
		err = store.Save(m.path, db)
	}
	if err != nil {
		m.err = err
		return m
	}
	m.db = db
	m.sessionEdit = nil
	m.err, m.message = nil, "sesion actualizada"
	return m
}

func parseSessionDate(value string, original time.Time) (time.Time, error) {
	// Preserve subsecond precision and timezone when a field is unchanged.
	if value == original.Local().Format(sessionDateFormat) {
		return original, nil
	}
	parsed, err := time.ParseInLocation(sessionDateFormat, value, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("usa fecha y hora: AAAA-MM-DD HH:MM:SS")
	}
	return parsed, nil
}
