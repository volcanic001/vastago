package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestFocusHeatmapIsWeekGridAndFitsWidth(t *testing.T) {
	loc := time.FixedZone("local", -6*3600)
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, loc)
	activity := make([]store.DayActivity, 40)
	for i := range activity {
		activity[i].Date = start.AddDate(0, 0, i)
		if i%3 == 0 {
			activity[i].Duration = time.Duration(i+1) * time.Minute
		}
	}
	for _, width := range []int{8, 16, 80} {
		lines := focusHeatmap(width, activity)
		if len(lines) != 8 {
			t.Fatalf("width %d: got %d lines", width, len(lines))
		}
		for _, line := range lines {
			if got := len([]rune(ansi.Strip(line))); got > width {
				t.Fatalf("width %d: line width %d: %q", width, got, ansi.Strip(line))
			}
		}
	}
}

func TestFocusHeatmapHandlesNoActivity(t *testing.T) {
	loc := time.FixedZone("local", 0)
	lines := focusHeatmap(20, []store.DayActivity{{Date: time.Date(2026, 1, 1, 0, 0, 0, 0, loc)}})
	if !strings.Contains(ansi.Strip(strings.Join(lines, "")), "·") {
		t.Fatal("missing empty activity dots")
	}
}
