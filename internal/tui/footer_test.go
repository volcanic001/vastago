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

func TestSessionEditorFooterShowsEditingKeys(t *testing.T) {
	model := sessionFixture(t)
	model = pressSession(model, rune(101), "e")
	footer := strings.TrimSpace(ansi.Strip(model.footer(120)))
	want := "[↑/↓/tab] campo · [ctrl+s] guardar · [ctrl+u] limpiar · [esc] cancelar"
	if footer != want {
		t.Fatalf("editor footer = %q, want %q", footer, want)
	}
}

func TestSessionsFooterIncludesRepeat(t *testing.T) {
	model := Model{db: &store.Database{}, screen: sessionsScreen, width: 100}
	footer := ansi.Strip(model.footer(model.width))
	if !strings.Contains(footer, "[r] repetir") {
		t.Fatalf("sessions footer = %q", footer)
	}
}

func TestMetricsFooterUsesTwoRowsWhenNeeded(t *testing.T) {
	model := Model{db: &store.Database{}, screen: metricsScreen, width: 80}
	footer := ansi.Strip(model.footer(model.width))
	if strings.Count(footer, "\n") != 2 {
		t.Fatalf("metrics footer should use two rows when needed: %q", footer)
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
		if !strings.Contains(footer, "[y] borrar    [n/esc] cancelar") {
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

func TestShortcutMenuWrapsWholeItems(t *testing.T) {
	items := []shortcut{{key: "a", action: "alfa"}, {key: "b", action: "beta"}, {key: "c", action: "gama"}}
	got := ansi.Strip(shortcutMenu(20, " · ", items))
	want := "[a] alfa\n[b] beta · [c] gama"
	if got != want {
		t.Fatalf("shortcut menu = %q, want %q", got, want)
	}
}

func TestConfirmationShortcutsWrapAtNarrowWidth(t *testing.T) {
	items := []shortcut{{key: "y", action: "borrar", tone: shortcutDanger}, {key: "n/esc", action: "cancelar"}}
	got := ansi.Strip(shortcutMenu(16, "    ", items))
	want := "[y] borrar\n[n/esc] cancelar"
	if got != want {
		t.Fatalf("confirmation shortcuts = %q, want %q", got, want)
	}
}

func TestShortcutMenuKeepsFirstRowCompact(t *testing.T) {
	items := []shortcut{{key: "a", action: "a"}, {key: "b", action: "b"}, {key: "c", action: "c"}, {key: "d", action: "d"}, {key: "e", action: "e"}}
	got := ansi.Strip(shortcutMenu(29, " · ", items))
	want := "[a] a\n[b] b · [c] c · [d] d · [e] e"
	if got != want {
		t.Fatalf("shortcut menu = %q, want compact first row %q", got, want)
	}
}
