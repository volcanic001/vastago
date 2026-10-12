package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

func habitHistoryFixture() Model {
	now := time.Date(2026, 6, 15, 9, 0, 0, 0, time.Local)
	db := &store.Database{Habits: []store.Habit{
		{ID: "1", Name: "meditar", CreatedAt: time.Date(2026, 3, 10, 0, 0, 0, 0, time.Local), Completions: []string{"2026-03-12", "2026-06-01"}},
	}}
	return Model{db: db, now: now, width: 90, height: 30, screen: habitsScreen}
}

func historyKey(m Model, code rune, text string) Model {
	updated, _ := m.handleKey(keyMessage(code, text))
	return updated.(Model)
}

func TestHabitHistoryOpenClose(t *testing.T) {
	m := habitHistoryFixture()
	m = historyKey(m, 'i', "i")
	if !m.habitHistory || m.habitHistoryYear != 2026 {
		t.Fatalf("i did not open history at current year: %+v", m)
	}
	m = historyKey(m, 'i', "i")
	if m.habitHistory {
		t.Fatal("i did not close history")
	}

	m = historyKey(m, 'i', "i")
	m.habitHistoryScroll, m.habitHistoryScrollX = 3, 3
	updated, _ := m.handleKey(tea.KeyPressMsg(tea.Key{Code: tea.KeyEsc}))
	m = updated.(Model)
	if m.habitHistory {
		t.Fatal("esc did not close history")
	}
}

func TestHabitHistoryYearNavigationResetsScroll(t *testing.T) {
	m := historyKey(habitHistoryFixture(), 'i', "i")
	m.habitHistoryScroll, m.habitHistoryScrollX = 4, 4
	m = historyKey(m, ']', "]")
	if m.habitHistoryYear != 2027 || m.habitHistoryScroll != 0 || m.habitHistoryScrollX != 0 {
		t.Fatalf("] did not advance year and reset scroll: %+v", m)
	}
	m = historyKey(m, '[', "[")
	m = historyKey(m, '[', "[")
	if m.habitHistoryYear != 2025 {
		t.Fatalf("[ did not go back a year: %+v", m)
	}
}

func TestHabitHistoryScrollClamped(t *testing.T) {
	m := historyKey(habitHistoryFixture(), 'i', "i")
	m.width = 16
	for i := 0; i < 500; i++ {
		m = historyKey(m, 'l', "l")
	}
	_, _, total := habitYearLayout(m.habitHistoryYear)
	gridWidth := max(1, m.width-2)
	if want := max(0, total-gridWidth); m.habitHistoryScrollX != want {
		t.Fatalf("horizontal scroll = %d, want clamp at %d", m.habitHistoryScrollX, want)
	}
	for i := 0; i < 500; i++ {
		m = historyKey(m, 'h', "h")
	}
	if m.habitHistoryScrollX != 0 {
		t.Fatalf("horizontal scroll did not clamp at 0: %d", m.habitHistoryScrollX)
	}

	m.height = 10
	for i := 0; i < 500; i++ {
		m = historyKey(m, 'j', "j")
	}
	_, limit := m.habitHistoryPage(m.width)
	if m.habitHistoryScroll != limit {
		t.Fatalf("vertical scroll = %d, want clamp at %d", m.habitHistoryScroll, limit)
	}
}

func TestHabitHistorySwallowsGlobalKeys(t *testing.T) {
	m := historyKey(habitHistoryFixture(), 'i', "i")
	updated, _ := m.handleKey(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if updated.(Model).screen != habitsScreen {
		t.Fatal("tab escaped the history overlay")
	}
	updated, _ = m.handleKey(keyMessage('q', "q"))
	if _, ok := updated.(Model); !ok {
		t.Fatal("q quit the program while history was open")
	}
}

func TestHabitYearLayoutMonthsAreNotUniform(t *testing.T) {
	starts, labels, total := habitYearLayout(2026)
	if len(starts) != 12 || len(labels) != 12 {
		t.Fatalf("expected 12 months, got %d/%d", len(starts), len(labels))
	}
	sum, distinct := 0, map[int]bool{}
	for month := time.January; month <= time.December; month++ {
		weeks := weeksInMonth(2026, month)
		sum += weeks
		distinct[weeks] = true
	}
	if total != sum+11 {
		t.Fatalf("total width = %d, want %d (11 gap columns between 12 months)", total, sum+11)
	}
	if len(distinct) < 2 {
		t.Fatal("expected months to span different week counts, like the reference calendar layout")
	}
}

func TestHabitYearGridShowsFullCalendarRegardlessOfCreationDate(t *testing.T) {
	// Matches reference contribution calendars (GitHub-style): the grid stays
	// visually whole for the year even before the habit existed, instead of
	// punching blank holes in it.
	habit := store.Habit{
		CreatedAt:   time.Date(2026, 9, 16, 0, 0, 0, 0, time.Local),
		Completions: []string{"2026-03-12"},
	}
	_, _, total := habitYearLayout(2026)
	grid := habitYearGrid(habit, 2026, total)

	beforeCreation := time.Date(2026, 1, 15, 0, 0, 0, 0, time.Local)
	if cell := cellAt(grid, 2026, beforeCreation); cell != cellMuted {
		t.Fatalf("day before CreatedAt = %d, want cellMuted", cell)
	}
	marked := time.Date(2026, 3, 12, 0, 0, 0, 0, time.Local)
	if cell := cellAt(grid, 2026, marked); cell != cellMarked {
		t.Fatalf("completed day = %d, want cellMarked", cell)
	}
	unmarked := time.Date(2026, 4, 1, 0, 0, 0, 0, time.Local)
	if cell := cellAt(grid, 2026, unmarked); cell != cellMuted {
		t.Fatalf("uncompleted day = %d, want cellMuted", cell)
	}
}

func TestHabitYearGridBlanksOutsideMonthBounds(t *testing.T) {
	// The only legitimate blank cells are the padding days that spill into a
	// month's first/last week but belong to a neighboring month, plus the
	// explicit gap column between months.
	habit := store.Habit{CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)}
	_, _, total := habitYearLayout(2026)
	grid := habitYearGrid(habit, 2026, total)

	// Jan 1, 2026 is a Thursday, so Mon-Wed of its first week belong to
	// December 2025 and must stay blank.
	padding := time.Date(2025, 12, 29, 0, 0, 0, 0, time.Local)
	col := int(padding.Sub(weekStart(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))).Hours()) / (7 * 24)
	dayIndex := (int(padding.Weekday()) + 6) % 7
	if cell := grid[dayIndex][col]; cell != cellBlank {
		t.Fatalf("padding day before month start = %d, want cellBlank", cell)
	}
}

// cellAt locates the grid column for a date by walking the same month/week
// layout habitYearGrid uses, so the test stays correct if month widths change.
func cellAt(grid [7][]byte, year int, day time.Time) byte {
	col := 0
	for month := time.January; month < day.Month(); month++ {
		col += weeksInMonth(year, month) + 1
	}
	monthStart := time.Date(year, day.Month(), 1, 0, 0, 0, 0, time.Local)
	col += int(day.Sub(weekStart(monthStart)).Hours()) / (7 * 24)
	dayIndex := (int(day.Weekday()) + 6) % 7
	return grid[dayIndex][col]
}

func TestHabitHistoryFitsNarrowTerminal(t *testing.T) {
	db := &store.Database{}
	for index := 0; index < 8; index++ {
		db.Habits = append(db.Habits, store.Habit{
			ID: string(rune('a' + index)), Name: "Un habito con nombre bastante largo",
			CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), Completions: []string{"2026-06-15"},
		})
	}
	m := Model{db: db, now: time.Date(2026, 6, 15, 9, 0, 0, 0, time.Local), width: 18, height: 14, screen: habitsScreen}
	m = historyKey(m, 'i', "i")
	content := m.View().Content
	for number, line := range strings.Split(content, "\n") {
		if got := lipgloss.Width(line); got > m.width {
			t.Fatalf("line %d width = %d; line %q", number+1, got, line)
		}
	}
	if got := len(strings.Split(content, "\n")); got > m.height {
		t.Fatalf("height = %d, want <= %d", got, m.height)
	}
}

func TestHabitHistoryEmptyState(t *testing.T) {
	m := Model{db: &store.Database{}, now: time.Now(), screen: habitsScreen, width: 40, height: 20}
	updated, _ := m.handleKey(keyMessage('i', "i"))
	if updated.(Model).habitHistory {
		t.Fatal("i opened history with zero habits")
	}
}
