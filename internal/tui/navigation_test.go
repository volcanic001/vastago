package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/volcanic001/vastago/internal/store"
)

func TestNavigationClearsTransientFeedback(t *testing.T) {
	m := Model{db: &store.Database{}, screen: sessionsScreen, message: "terminada: Trabajo", err: errors.New("error anterior")}
	updated, _ := m.handleKey(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	m = updated.(Model)
	if m.screen != todosScreen || m.message != "" || m.err != nil {
		t.Fatalf("feedback after leaving sessions = screen %s, message %q, err %v", m.screen.name(), m.message, m.err)
	}
	updated, _ = m.handleKey(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}))
	m = updated.(Model)
	if m.screen != sessionsScreen || m.message != "" || m.err != nil {
		t.Fatalf("feedback after returning to sessions = message %q, err %v", m.message, m.err)
	}
}

func TestForwardNavigationKeysSelectNextScreen(t *testing.T) {
	for _, key := range []tea.Key{
		{Code: tea.KeyTab},
		{Code: tea.KeyRight},
		{Code: 'l', Text: "l"},
	} {
		updated, _ := (Model{db: &store.Database{}}).handleKey(tea.KeyPressMsg(key))
		if got := updated.(Model).screen; got != todosScreen {
			t.Errorf("key %q selected screen %d, want %d", key.String(), got, todosScreen)
		}
	}
}

func TestReverseNavigationWrapsToMetrics(t *testing.T) {
	for _, key := range []tea.Key{
		{Code: tea.KeyTab, Mod: tea.ModShift},
		{Code: tea.KeyLeft},
		{Code: 'h', Text: "h"},
	} {
		updated, _ := (Model{db: &store.Database{}}).handleKey(tea.KeyPressMsg(key))
		if got := updated.(Model).screen; got != metricsScreen {
			t.Errorf("key %q selected screen %d, want %d", key.String(), got, metricsScreen)
		}
	}
}

func TestNumberKeysSelectScreens(t *testing.T) {
	want := []screen{sessionsScreen, todosScreen, habitsScreen, metricsScreen}
	for index, expected := range want {
		key := rune('1' + index)
		updated, _ := (Model{db: &store.Database{}}).handleKey(tea.KeyPressMsg(tea.Key{Code: key, Text: string(key)}))
		if got := updated.(Model).screen; got != expected {
			t.Errorf("key %q selected screen %d, want %d", key, got, expected)
		}
	}
}

func TestForwardNavigationWrapsToSessions(t *testing.T) {
	if got := (Model{screen: metricsScreen}).nextScreen().screen; got != sessionsScreen {
		t.Fatalf("next screen = %d, want %d", got, sessionsScreen)
	}
}
