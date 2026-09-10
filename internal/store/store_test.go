package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStartStopAndPersist(t *testing.T) {
	now := time.Date(2026, time.September, 9, 9, 0, 0, 0, time.Local)
	db := &Database{}
	if _, err := db.Start(now, "Trabajo profundo", "sin distracciones"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if _, err := db.Start(now, "Otra", ""); err != ErrActiveEntry {
		t.Fatalf("second Start() error = %v, want %v", err, ErrActiveEntry)
	}
	finished, err := db.Stop(now.Add(90*time.Minute), "terminado")
	if err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if got := finished.Duration(now.Add(3 * time.Hour)); got != 90*time.Minute {
		t.Fatalf("Duration() = %v, want 90m", got)
	}

	path := filepath.Join(t.TempDir(), "nested", "store.json")
	if err := Save(path, db); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Entries) != 1 || loaded.Entries[0].Task != "Trabajo profundo" {
		t.Fatalf("Load() entries = %#v", loaded.Entries)
	}
}

func TestTotalsAreSorted(t *testing.T) {
	now := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.UTC)
	endA := now.Add(-time.Hour)
	endB := now
	entries := []Entry{
		{Task: "Lectura", Start: now.Add(-2 * time.Hour), End: &endA},
		{Task: "Construir", Start: now.Add(-3 * time.Hour), End: &endB},
	}
	totals := Totals(entries, now)
	if len(totals) != 2 || totals[0].Task != "Construir" {
		t.Fatalf("Totals() = %#v", totals)
	}
}

func TestTodoLifecycleAndPersistence(t *testing.T) {
	now := time.Date(2026, time.September, 10, 8, 0, 0, 0, time.Local)
	db := &Database{}
	todo, err := db.AddTodo(now, "  Preparar el diseño  ")
	if err != nil {
		t.Fatalf("AddTodo() error = %v", err)
	}
	if todo.Title != "Preparar el diseño" || todo.Completed() {
		t.Fatalf("AddTodo() = %#v", todo)
	}
	if _, err := db.RenameTodo(todo.ID, "Diseñar pendientes"); err != nil {
		t.Fatalf("RenameTodo() error = %v", err)
	}
	if _, err := db.ToggleTodo(now.Add(time.Hour), todo.ID); err != nil {
		t.Fatalf("ToggleTodo() error = %v", err)
	}
	if !db.Todos[0].Completed() {
		t.Fatal("ToggleTodo() did not complete todo")
	}

	path := filepath.Join(t.TempDir(), "store.json")
	if err := Save(path, db); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Todos) != 1 || loaded.Todos[0].Title != "Diseñar pendientes" || !loaded.Todos[0].Completed() {
		t.Fatalf("Load() todos = %#v", loaded.Todos)
	}
	if err := loaded.DeleteTodo(todo.ID); err != nil {
		t.Fatalf("DeleteTodo() error = %v", err)
	}
	if len(loaded.Todos) != 0 {
		t.Fatalf("DeleteTodo() left %d todos", len(loaded.Todos))
	}
}

func TestExistingDatabaseWithoutTodosRemainsCompatible(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	contents := []byte(`{"entries":[]}`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	db, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(db.Entries) != 0 || len(db.Todos) != 0 {
		t.Fatalf("Load() database = %#v", db)
	}
}
