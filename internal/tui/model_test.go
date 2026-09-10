package tui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

func TestResponsiveViewsFitTerminalWidth(t *testing.T) {
	now := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.Local)
	end := now.Add(-time.Hour)
	db := &store.Database{}
	for index := 0; index < 18; index++ {
		db.Entries = append(db.Entries, store.Entry{
			Task:  "Construir algo que merezca crecer",
			Start: now.Add(-time.Duration(index+3) * time.Hour),
			End:   &end,
		})
	}

	for _, size := range []struct {
		width  int
		height int
	}{
		{16, 12},
		{32, 16},
		{60, 20},
		{96, 30},
		{140, 40},
	} {
		for page := 0; page < 2; page++ {
			model := Model{db: db, now: now, width: size.width, height: size.height, page: page}
			content := model.View().Content
			for number, line := range strings.Split(content, "\n") {
				if got := lipgloss.Width(line); got > size.width {
					t.Fatalf("size %dx%d page %d line %d width = %d; line %q", size.width, size.height, page, number+1, got, line)
				}
			}
			if got := len(strings.Split(content, "\n")); got > size.height {
				t.Fatalf("size %dx%d page %d height = %d", size.width, size.height, page, got)
			}
		}
	}
}

func TestTinyDashboardShowsEssentialInformation(t *testing.T) {
	now := time.Date(2026, time.September, 9, 12, 0, 0, 0, time.Local)
	model := Model{db: &store.Database{}, now: now, width: 32, height: 12}
	content := model.View().Content
	for _, expected := range []string{"VASTAGO", "EN PAUSA", "hoy", "7 dias"} {
		if !strings.Contains(content, expected) {
			t.Fatalf("View() does not contain %q: %q", expected, content)
		}
	}
}
