package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestListFootersDescribeBothNavigationMethods(t *testing.T) {
	for _, current := range []screen{sessionsScreen, todosScreen, habitsScreen, metricsScreen} {
		model := Model{db: &store.Database{}, screen: current, width: 100}
		footer := ansi.Strip(model.footer(model.width))
		if !strings.Contains(footer, "j/k") || !strings.Contains(footer, "↑↓") {
			t.Fatalf("screen %s footer lacks navigation keys: %q", current.name(), footer)
		}
	}
}

func TestTodoAndHabitConfirmationReplaceNormalActions(t *testing.T) {
	for _, current := range []screen{todosScreen, habitsScreen} {
		model := Model{db: &store.Database{}, screen: current, width: 100}
		if current == todosScreen {
			model.confirmTodo = true
		} else {
			model.confirmHabit = true
		}
		footer := ansi.Strip(model.footer(model.width))
		if !strings.Contains(footer, "y confirmar · n o esc cancelar") {
			t.Fatalf("screen %s confirmation footer = %q", current.name(), footer)
		}
		if strings.Contains(footer, "space ") {
			t.Fatalf("screen %s showed normal actions during confirmation: %q", current.name(), footer)
		}
	}
}

func TestListFootersStayWithinNarrowTerminal(t *testing.T) {
	for _, current := range []screen{todosScreen, habitsScreen} {
		model := Model{db: &store.Database{}, screen: current, width: 16}
		if got := len([]rune(ansi.Strip(model.footer(model.width)))); got == 0 {
			t.Fatalf("screen %s footer is empty", current.name())
		}
	}
}
