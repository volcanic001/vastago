package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestSelectedHabitLeavesCalendarInSeparateColumn(t *testing.T) {
	now := time.Date(2026, time.September, 10, 9, 0, 0, 0, time.Local)
	habit := store.Habit{Name: "prayer", Completions: []string{"2026-09-10"}}
	for _, width := range []int{16, 52, 80} {
		labelWidth, gap, _, visible := habitCalendarLayout(width)
		if !visible {
			t.Fatalf("calendar hidden at width %d", width)
		}
		line := habitListLine(habit, now, true, true, width)
		plain := ansi.Strip(line)
		if got := len([]rune(plain)); got != width {
			t.Fatalf("width %d: rendered width %d: %q", width, got, plain)
		}
		calendar := string([]rune(plain)[labelWidth+gap:])
		if len([]rune(calendar)) != map[bool]int{true: 13, false: 7}[width >= compactBreakpoint] {
			t.Fatalf("width %d: calendar column = %q", width, calendar)
		}
		selected := selectionStyle.Width(labelWidth).Render("● " + trimToWidth(habit.Name, max(1, labelWidth-2)))
		if !strings.HasPrefix(line, selected) {
			t.Fatalf("width %d: selection does not end before calendar", width)
		}
	}
}
