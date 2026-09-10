package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/volcanic001/vastago/internal/store"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func sessionFixture(t *testing.T) Model {
	t.Helper()
	start := time.Date(2026, 9, 10, 9, 0, 0, 123, time.Local)
	end := start.Add(time.Hour)
	db := &store.Database{
		Entries: []store.Entry{{ID: "old", Task: "Lectura", Start: start, End: &end}, {ID: "new", Task: "Trabajo", Start: start, End: &end}},
		Todos:   []store.Todo{{ID: "todo", Title: "Conservar"}},
		Habits:  []store.Habit{{ID: "habit", Name: "Conservar"}},
	}
	path := filepath.Join(t.TempDir(), "store.json")
	if err := store.Save(path, db); err != nil {
		t.Fatal(err)
	}
	return Model{db: db, path: path, now: end, width: 80, height: 24, screen: sessionsScreen}
}

func pressSession(m Model, code rune, text string) Model {
	updated, _ := m.handleKey(keyMessage(code, text))
	return updated.(Model)
}

func TestSessionEditAndDeleteKeyboard(t *testing.T) {
	m := sessionFixture(t)
	m = pressSession(m, 'j', "j")
	m = pressSession(m, 'e', "e")
	if m.sessionEdit == nil || m.sessionEdit.id != "old" {
		t.Fatal("wrong selection")
	}
	m = pressSession(m, 'u', "") // no text should leave input unchanged
	m.sessionEdit.values[0] = "Lectura revisada"
	m.sessionEdit.values[1] = "Nota nueva"
	for i := 0; i < 4; i++ {
		m = pressSession(m, tea.KeyEnter, "")
	}
	if m.err != nil || m.sessionEdit != nil {
		t.Fatalf("save: %v", m.err)
	}
	db, err := store.Load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if db.Entries[0].Task != "Lectura revisada" || db.Entries[0].Note != "Nota nueva" {
		t.Fatal("edit not persisted")
	}
	if db.Entries[0].Start.Nanosecond() != 123 {
		t.Fatal("unchanged timestamp lost precision")
	}
	m = pressSession(m, 'd', "d")
	m = pressSession(m, tea.KeyEscape, "")
	if len(m.db.Entries) != 2 {
		t.Fatal("cancel deleted data")
	}
	m = pressSession(m, 'd', "d")
	m = pressSession(m, 'y', "y")
	db, err = store.Load(m.path)
	if err != nil {
		t.Fatal(err)
	}
	if len(db.Entries) != 1 || db.Entries[0].ID != "new" || len(db.Todos) != 1 || len(db.Habits) != 1 {
		t.Fatal("wrong data after deletion")
	}
}

func TestSessionFailedSaveKeepsState(t *testing.T) {
	for _, operation := range []string{"edit", "delete"} {
		t.Run(operation, func(t *testing.T) {
			m := sessionFixture(t)
			before := append([]store.Entry(nil), m.db.Entries...)
			// An existing file used as a parent deterministically prevents saving.
			m.path = filepath.Join(m.path, "cannot-save.json")
			if operation == "edit" {
				m = pressSession(m, 'e', "e")
				m.sessionEdit.values[0] = "Must not appear"
				m.sessionEdit.field = 3
				m = pressSession(m, tea.KeyEnter, "")
			} else {
				m = pressSession(m, 'd', "d")
				m = pressSession(m, 'y', "y")
			}
			if m.err == nil {
				t.Fatal("expected save error")
			}
			if !reflect.DeepEqual(before, m.db.Entries) {
				t.Fatal("failed save mutated entries")
			}
		})
	}
}

func TestSessionViewsFitAndReachOldest(t *testing.T) {
	m := sessionFixture(t)
	for i := 0; i < 40; i++ {
		entry := m.db.Entries[0]
		entry.ID = string(rune(100 + i))
		m.db.Entries = append(m.db.Entries, entry)
	}
	for i := 0; i < 50; i++ {
		m = pressSession(m, tea.KeyDown, "")
	}
	if m.selectedSession != len(m.db.Entries)-1 {
		t.Fatal("cannot reach oldest session")
	}
	for _, size := range [][2]int{{16, 12}, {32, 16}, {60, 20}, {96, 30}} {
		for _, mode := range []string{"list", "edit", "delete"} {
			view := m
			view.width, view.height = size[0], size[1]
			if mode == "edit" {
				view = pressSession(view, 'e', "e")
			}
			if mode == "delete" {
				view = pressSession(view, 'd', "d")
			}
			content := view.View().Content
			if len(strings.Split(content, "\n")) > view.height {
				t.Fatalf("%v %s too tall", size, mode)
			}
			for _, line := range strings.Split(content, "\n") {
				if lipgloss.Width(line) > view.width {
					t.Fatalf("%v %s too wide", size, mode)
				}
			}
		}
	}
}
