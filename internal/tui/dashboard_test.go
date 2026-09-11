package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestDashboardSummarizesTodayAndQuickActions(t *testing.T) {
	now := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.Local)
	end := now.Add(-time.Hour)
	db := &store.Database{
		Entries: []store.Entry{{Task: "Lectura", Start: now.Add(-2 * time.Hour), End: &end}},
		Todos: []store.Todo{
			{ID: "open", Title: "Escribir"},
			{ID: "done", Title: "Revisar", CompletedAt: &end},
		},
		Habits: []store.Habit{
			{ID: "done", Name: "Caminar", Completions: []string{now.Format("2006-01-02")}},
			{ID: "open", Name: "Leer"},
		},
	}
	model := Model{db: db, now: now, width: 80, height: 24}
	content := ansi.Strip(model.View().Content)
	for _, want := range []string{"RITMO", "Semana", "HOY", "Habitos", "1/2 hoy", "Pendientes", "1 abiertos", "[n] iniciar"} {
		if !strings.Contains(content, want) {
			t.Fatalf("dashboard missing %q: %s", want, content)
		}
	}
}

func TestTinyDashboardKeepsTodayStatus(t *testing.T) {
	now := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.Local)
	db := &store.Database{
		Todos:  []store.Todo{{ID: "open"}},
		Habits: []store.Habit{{ID: "done", Completions: []string{now.Format("2006-01-02")}}},
	}
	model := Model{db: db, now: now, width: 32, height: 12}
	content := ansi.Strip(model.View().Content)
	for _, want := range []string{"habitos 1/1", "pendientes 1"} {
		if !strings.Contains(content, want) {
			t.Fatalf("tiny dashboard missing %q: %s", want, content)
		}
	}
}
