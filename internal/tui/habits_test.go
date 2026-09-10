package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

func TestHabitKeyboardLifecyclePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	now := time.Date(2026, time.September, 10, 9, 0, 0, 0, time.Local)
	model := Model{path: path, db: &store.Database{}, now: now, screen: habitsScreen}

	updated, _ := model.handleKey(keyMessage('n', "n"))
	model = updated.(Model)
	if !model.inputMode || model.inputAction != inputHabitNew {
		t.Fatal("n did not open new habit input")
	}
	model.input = []rune("Leer 20 minutos")
	updated, _ = model.handleKey(keyMessage(tea.KeyEnter, ""))
	model = updated.(Model)
	if len(model.db.Habits) != 1 || model.db.Habits[0].Name != "Leer 20 minutos" {
		t.Fatalf("created habits = %#v", model.db.Habits)
	}

	updated, _ = model.handleKey(keyMessage(tea.KeySpace, ""))
	model = updated.(Model)
	if !model.db.Habits[0].CompletedOn(now) {
		t.Fatal("space did not mark selected habit for today")
	}

	updated, _ = model.handleKey(keyMessage('e', "e"))
	model = updated.(Model)
	model.input = []rune("Lectura diaria")
	updated, _ = model.handleKey(keyMessage(tea.KeyEnter, ""))
	model = updated.(Model)
	if model.db.Habits[0].Name != "Lectura diaria" {
		t.Fatalf("edited name = %q", model.db.Habits[0].Name)
	}

	updated, _ = model.handleKey(keyMessage('d', "d"))
	model = updated.(Model)
	if !model.confirmHabit {
		t.Fatal("d did not request habit deletion confirmation")
	}
	updated, _ = model.handleKey(keyMessage('y', "y"))
	model = updated.(Model)
	if len(model.db.Habits) != 0 {
		t.Fatalf("delete left %d habits", len(model.db.Habits))
	}

	loaded, err := store.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Habits) != 0 {
		t.Fatalf("persisted habits = %#v", loaded.Habits)
	}
}

func TestHabitListFitsSmallTerminal(t *testing.T) {
	now := time.Date(2026, time.September, 10, 9, 0, 0, 0, time.Local)
	db := &store.Database{}
	for index := 0; index < 20; index++ {
		db.Habits = append(db.Habits, store.Habit{
			ID:          string(rune('a' + index)),
			Name:        "Habito con un nombre bastante largo",
			Completions: []string{now.Format("2006-01-02")},
		})
	}
	model := Model{db: db, now: now, width: 16, height: 12, screen: habitsScreen, selectedHabit: 19}
	content := model.View().Content
	for number, line := range strings.Split(content, "\n") {
		if got := lipgloss.Width(line); got > model.width {
			t.Fatalf("line %d width = %d; line %q", number+1, got, line)
		}
	}
	if got := len(strings.Split(content, "\n")); got > model.height {
		t.Fatalf("height = %d, want <= %d", got, model.height)
	}
}
