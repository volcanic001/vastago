package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestHabitWeekGridStartsMonday(t *testing.T) {
	now := time.Date(2026, time.September, 10, 9, 0, 0, 0, time.Local) // Thursday
	habit := store.Habit{Completions: []string{"2026-09-07", "2026-09-10"}}
	grid := ansi.Strip(habitWeekGrid(habit, now, true))
	if grid != "● · · ● · · ·" {
		t.Fatalf("week grid = %q", grid)
	}
}

func TestHabitsScreenShowsDailyAndWeeklyProgress(t *testing.T) {
	now := time.Date(2026, time.September, 10, 9, 0, 0, 0, time.Local)
	db := &store.Database{Habits: []store.Habit{
		{ID: "one", Name: "Caminar", CreatedAt: now.AddDate(0, 0, -3), Completions: []string{"2026-09-08", "2026-09-10"}},
		{ID: "two", Name: "Leer", CreatedAt: now.AddDate(0, 0, -3), Completions: []string{"2026-09-09"}},
	}}
	model := Model{db: db, now: now, width: 80, height: 24, screen: habitsScreen}
	content := ansi.Strip(model.View().Content)
	for _, want := range []string{"HABITOS · 1/2 hoy · 3/8 semana", "L M X J V S D", "Caminar", "● · ●"} {
		if !strings.Contains(content, want) {
			t.Fatalf("habit screen missing %q: %q", want, content)
		}
	}
}
