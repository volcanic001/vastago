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

func TestTodoKeyboardLifecyclePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "store.json")
	model := Model{path: path, db: &store.Database{}, now: time.Now(), screen: todosScreen}

	updated, _ := model.handleKey(keyMessage('n', "n"))
	model = updated.(Model)
	if !model.inputMode || model.inputAction != inputTodoNew {
		t.Fatal("n did not open new todo input")
	}
	model.input = []rune("Preparar métricas")
	updated, _ = model.handleKey(keyMessage(tea.KeyEnter, ""))
	model = updated.(Model)
	if len(model.db.Todos) != 1 || model.db.Todos[0].Title != "Preparar métricas" {
		t.Fatalf("created todos = %#v", model.db.Todos)
	}

	updated, _ = model.handleKey(keyMessage(tea.KeySpace, ""))
	model = updated.(Model)
	if !model.db.Todos[0].Completed() {
		t.Fatal("space did not complete selected todo")
	}

	updated, _ = model.handleKey(keyMessage('e', "e"))
	model = updated.(Model)
	model.input = []rune("Diseñar métricas")
	updated, _ = model.handleKey(keyMessage(tea.KeyEnter, ""))
	model = updated.(Model)
	if model.db.Todos[0].Title != "Diseñar métricas" {
		t.Fatalf("edited title = %q", model.db.Todos[0].Title)
	}

	updated, _ = model.handleKey(keyMessage('d', "d"))
	model = updated.(Model)
	if !model.confirmTodo {
		t.Fatal("d did not request confirmation")
	}
	updated, _ = model.handleKey(keyMessage('y', "y"))
	model = updated.(Model)
	if len(model.db.Todos) != 0 {
		t.Fatalf("delete left %d todos", len(model.db.Todos))
	}

	loaded, err := store.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Todos) != 0 {
		t.Fatalf("persisted todos = %#v", loaded.Todos)
	}
}

func TestTodoSelectionUsesArrowAndVimKeys(t *testing.T) {
	model := Model{db: &store.Database{Todos: []store.Todo{{ID: "1"}, {ID: "2"}}}, screen: todosScreen}
	updated, _ := model.handleKey(keyMessage('j', "j"))
	model = updated.(Model)
	if model.selectedTodo != 1 {
		t.Fatalf("j selected %d, want 1", model.selectedTodo)
	}
	updated, _ = model.handleKey(keyMessage(tea.KeyUp, ""))
	if got := updated.(Model).selectedTodo; got != 0 {
		t.Fatalf("up selected %d, want 0", got)
	}
}

func TestTodoListFitsSmallTerminal(t *testing.T) {
	db := &store.Database{}
	for index := 0; index < 20; index++ {
		db.Todos = append(db.Todos, store.Todo{ID: string(rune('a' + index)), Title: "Pendiente con un nombre bastante largo"})
	}
	model := Model{db: db, now: time.Now(), width: 16, height: 12, screen: todosScreen, selectedTodo: 19}
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

func keyMessage(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code, Text: text})
}
