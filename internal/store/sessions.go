package store

import (
	"errors"
	"strings"
	"time"
)

var ErrSessionNotFound = errors.New("sesion no encontrada")

// ValidateSession checks the invariants for a session that is being stored.
func ValidateSession(start time.Time, end *time.Time, now time.Time) error {
	if start.IsZero() {
		return errors.New("inicio invalido")
	}
	if end == nil {
		return nil
	}
	if !start.Before(*end) {
		return errors.New("el fin debe ser posterior al inicio")
	}
	if end.After(now) {
		return errors.New("el fin no puede estar en el futuro")
	}
	localStart := start.In(time.Local)
	localEnd := end.In(time.Local)
	sy, sm, sd := localStart.Date()
	ey, em, ed := localEnd.Date()
	if sy != ey || sm != em || sd != ed {
		return errors.New("el inicio y el fin deben estar en el mismo dia")
	}
	return nil
}

// EditSession preserves identity and whether the session is active.
func (db *Database) EditSession(id, task, note string, start time.Time, end *time.Time, now time.Time) error {
	for i := range db.Entries {
		entry := &db.Entries[i]
		if entry.ID != id {
			continue
		}
		task = strings.TrimSpace(task)
		if task == "" {
			return errors.New("la tarea no puede estar vacia")
		}
		if (entry.End == nil) != (end == nil) {
			return errors.New("usa iniciar o terminar para cambiar el estado de la sesion")
		}
		if err := ValidateSession(start, end, now); err != nil {
			return err
		}
		entry.Task, entry.Note, entry.Start = task, strings.TrimSpace(note), start
		if end != nil {
			value := *end
			entry.End = &value
		}
		return nil
	}
	return ErrSessionNotFound
}

func (db *Database) DeleteSession(id string) error {
	for i := range db.Entries {
		if db.Entries[i].ID == id {
			db.Entries = append(db.Entries[:i], db.Entries[i+1:]...)
			return nil
		}
	}
	return ErrSessionNotFound
}
