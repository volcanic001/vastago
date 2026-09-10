package store

import (
	"errors"
	"strings"
	"time"
)

var ErrSessionNotFound = errors.New("sesion no encontrada")

// EditSession preserves identity and whether the session is active.
func (db *Database) EditSession(id, task, note string, start time.Time, end *time.Time) error {
	for i := range db.Entries {
		entry := &db.Entries[i]
		if entry.ID != id {
			continue
		}
		task = strings.TrimSpace(task)
		if task == "" {
			return errors.New("la tarea no puede estar vacia")
		}
		if start.IsZero() {
			return errors.New("inicio invalido")
		}
		if (entry.End == nil) != (end == nil) {
			return errors.New("usa iniciar o terminar para cambiar el estado de la sesion")
		}
		if end != nil && end.Before(start) {
			return errors.New("el fin no puede ser anterior al inicio")
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
