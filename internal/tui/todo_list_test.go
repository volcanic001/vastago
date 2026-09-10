package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestTodoListGroupsOpenBeforeCompleted(t *testing.T) {
	now := time.Date(2026, time.September, 10, 9, 0, 0, 0, time.Local)
	db := &store.Database{Todos: []store.Todo{
		{ID: "done", Title: "Archivado", CompletedAt: &now},
		{ID: "open-one", Title: "Primero"},
		{ID: "open-two", Title: "Segundo"},
	}}
	model := Model{db: db, now: now, width: 80, height: 24, screen: todosScreen, selectedTodo: 1}
	content := ansi.Strip(model.View().Content)
	for _, want := range []string{"PENDIENTES", "ABIERTOS · 2", "COMPLETADOS · 1"} {
		if !strings.Contains(content, want) {
			t.Fatalf("list missing %q: %q", want, content)
		}
	}
	if strings.Count(content, "ABIERTOS · 2") != 1 || strings.Count(content, "COMPLETADOS · 1") != 1 {
		t.Fatalf("section headings repeated: %q", content)
	}
	if strings.Index(content, "Primero") > strings.Index(content, "Archivado") {
		t.Fatalf("open todo was rendered after completed: %q", content)
	}

	model = model.moveTodoSelection(1)
	if model.selectedTodo != 2 {
		t.Fatalf("next selected %d, want second open todo", model.selectedTodo)
	}
	model = model.moveTodoSelection(1)
	if model.selectedTodo != 0 {
		t.Fatalf("next selected %d, want completed todo", model.selectedTodo)
	}
}
