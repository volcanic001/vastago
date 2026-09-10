package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNavigationKeysToggleScreens(t *testing.T) {
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

func TestReverseNavigationReturnsHome(t *testing.T) {
	for _, key := range []tea.Key{
		{Code: tea.KeyTab, Mod: tea.ModShift},
		{Code: tea.KeyLeft},
		{Code: 'h', Text: "h"},
	} {
		updated, _ := (Model{screen: sessionsScreen}).handleKey(tea.KeyPressMsg(key))
		if got := updated.(Model).screen; got != homeScreen {
			t.Errorf("key %q selected screen %d, want %d", key.String(), got, homeScreen)
		}
	}
}
