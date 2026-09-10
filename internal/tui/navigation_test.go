package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestForwardNavigationKeysSelectNextScreen(t *testing.T) {
	for _, key := range []tea.Key{
		{Code: tea.KeyTab},
		{Code: tea.KeyRight},
		{Code: 'l', Text: "l"},
	} {
		updated, _ := (Model{}).handleKey(tea.KeyPressMsg(key))
		if got := updated.(Model).screen; got != sessionsScreen {
			t.Errorf("key %q selected screen %d, want %d", key.String(), got, sessionsScreen)
		}
	}
}

func TestReverseNavigationWrapsToMetrics(t *testing.T) {
	for _, key := range []tea.Key{
		{Code: tea.KeyTab, Mod: tea.ModShift},
		{Code: tea.KeyLeft},
		{Code: 'h', Text: "h"},
	} {
		updated, _ := (Model{}).handleKey(tea.KeyPressMsg(key))
		if got := updated.(Model).screen; got != metricsScreen {
			t.Errorf("key %q selected screen %d, want %d", key.String(), got, metricsScreen)
		}
	}
}

func TestNumberKeysSelectScreens(t *testing.T) {
	want := []screen{homeScreen, sessionsScreen, todosScreen, habitsScreen, metricsScreen}
	for index, expected := range want {
		key := rune('1' + index)
		updated, _ := (Model{}).handleKey(tea.KeyPressMsg(tea.Key{Code: key, Text: string(key)}))
		if got := updated.(Model).screen; got != expected {
			t.Errorf("key %q selected screen %d, want %d", key, got, expected)
		}
	}
}

func TestForwardNavigationWrapsToHome(t *testing.T) {
	if got := (Model{screen: metricsScreen}).nextScreen().screen; got != homeScreen {
		t.Fatalf("next screen = %d, want %d", got, homeScreen)
	}
}
