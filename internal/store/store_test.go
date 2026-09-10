package store

import (
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
