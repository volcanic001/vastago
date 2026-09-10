package tui

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/volcanic001/vastago/internal/store"
)

func TestStoppingWithoutActiveSessionIsNeutral(t *testing.T) {
	model := Model{db: &store.Database{}, now: time.Now()}
	updated, _ := model.handleKey(tea.KeyPressMsg(tea.Key{Code: 'x', Text: "x"}))
	result := updated.(Model)
	if result.err != nil {
		t.Fatalf("handleKey(x) error = %v", result.err)
	}
	if result.message != "Sin sesión activa." {
		t.Fatalf("handleKey(x) message = %q", result.message)
	}
}
