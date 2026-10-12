package tui

import (
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

var spanishMonths = []string{"Ene", "Feb", "Mar", "Abr", "May", "Jun", "Jul", "Ago", "Sep", "Oct", "Nov", "Dic"}

const (
	cellBlank byte = iota
	cellMuted
	cellMarked
)

// handleHabitHistoryKey owns every key while the annual history is open,
// mirroring sessionDetailID's swallow-all behavior.
func (m Model) handleHabitHistoryKey(key tea.KeyPressMsg) (Model, bool) {
	switch key.String() {
	case "i", "esc":
		m.habitHistory = false
	case "[", "]":
		offset := 1
		if key.String() == "[" {
			offset = -1
		}
		m.habitHistoryYear += offset
		m.habitHistoryScroll, m.habitHistoryScrollX = 0, 0
	case "j", "down":
		_, limit := m.habitHistoryPage(max(1, m.width))
		m.habitHistoryScroll = min(m.habitHistoryScroll+1, limit)
	case "k", "up":
		m.habitHistoryScroll = max(0, m.habitHistoryScroll-1)
	case "l", "right":
		_, _, totalWidth := habitYearLayout(m.habitHistoryYear)
		maxOffsetX := max(0, totalWidth-max(1, m.width-2))
		m.habitHistoryScrollX = min(m.habitHistoryScrollX+1, maxOffsetX)
	case "h", "left":
		m.habitHistoryScrollX = max(0, m.habitHistoryScrollX-1)
	}
	m.message, m.err = "", nil
	return m, true
}

func (m Model) habitHistoryView(width int) string {
	content, _ := m.habitHistoryPage(width)
	return content
}

func (m Model) habitHistoryPage(width int) (string, int) {
	year := m.habitHistoryYear
	if year == 0 {
		year = m.now.Year()
	}
	period, err := store.NewPeriod(store.Year, time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local))
	if err != nil {
		return errorStyle.Render(trimToWidth(err.Error(), width)), 0
	}

	monthStarts, monthLabels, totalWidth := habitYearLayout(year)
	gridWidth := max(1, width-2)
	maxOffsetX := max(0, totalWidth-gridWidth)
	offsetX := min(m.habitHistoryScrollX, maxOffsetX)
	endX := min(totalWidth, offsetX+gridWidth)

	monthHeader := habitMonthHeader(monthStarts, monthLabels, offsetX, endX)

	var lines []string
	for index, habit := range m.db.Habits {
		if index > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, valueStyle.Render(trimToWidth(habit.Name, width)))
		grid := habitYearGrid(habit, year, totalWidth)
		lines = append(lines, renderHabitRows(grid, offsetX, endX)...)
	}
	if len(m.db.Habits) == 0 {
		lines = append(lines, mutedStyle.Render("Aun no hay habitos."))
	}

	wrapped := strings.Split(ansi.Wrap(strings.Join(lines, "\n"), max(1, width), ""), "\n")
	title := strings.Split(ansi.Wrap(periodLabel(period), max(1, width), ""), "\n")
	height := m.height
	if height <= 0 {
		height = 24
	}
	// The month header stays pinned above the scroll area so it's never lost
	// when scrolling vertically through several habits.
	available := max(1, height-lipgloss.Height(m.header(width))-lipgloss.Height(m.habitHistoryFooter(width))-2-len(title))
	offsetY := min(m.habitHistoryScroll, max(0, len(wrapped)-available))
	visible := wrapped[offsetY:min(len(wrapped), offsetY+available)]
	return "\n" + strings.Join(title, "\n") + "\n" + monthHeader + "\n" + strings.Join(visible, "\n"), max(0, len(wrapped)-available)
}

func (m Model) habitHistoryFooter(width int) string {
	items := []shortcut{{key: "h/l/←→", action: "meses", primary: true}, {key: "j/k/↑↓", action: "mover"}, {key: "[/]", action: "año", primary: true}, {key: "i/esc", action: "volver", primary: true}}
	rows := []string{shortcutMenu(width, " · ", items)}
	if m.err != nil {
		rows = append([]string{errorStyle.Render(trimToWidth("error: "+m.err.Error(), width))}, rows...)
	}
	return "\n" + strings.Join(rows, "\n")
}

// habitYearLayout computes where each month starts along the shared
// horizontal timeline (one column per calendar week, one blank column
// between months so they read as visually separated blocks).
func habitYearLayout(year int) (starts []int, labels []string, total int) {
	for month := time.January; month <= time.December; month++ {
		if month > time.January {
			total++ // gap column between months
		}
		starts = append(starts, total)
		labels = append(labels, spanishMonths[month-1])
		total += weeksInMonth(year, month)
	}
	return starts, labels, total
}

func weeksInMonth(year int, month time.Month) int {
	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	nextMonth := monthStart.AddDate(0, 1, 0)
	weeks := 0
	for week := weekStart(monthStart); week.Before(nextMonth); week = week.AddDate(0, 0, 7) {
		weeks++
	}
	return weeks
}

// habitYearGrid fills a [7][total]byte grid (Monday..Sunday rows) across the
// whole year's timeline for one habit, leaving month-gap columns blank.
// Months before the habit was created still get muted dots rather than
// blank cells, matching how reference contribution calendars (GitHub-style)
// keep the whole grid visually consistent instead of punching holes in it.
func habitYearGrid(habit store.Habit, year, total int) [7][]byte {
	marks := make(map[string]bool, len(habit.Completions))
	for _, completion := range habit.Completions {
		marks[completion] = true
	}

	var grid [7][]byte
	for day := range grid {
		grid[day] = make([]byte, 0, total)
	}
	for month := time.January; month <= time.December; month++ {
		if month > time.January {
			for day := range grid {
				grid[day] = append(grid[day], cellBlank)
			}
		}
		monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
		nextMonth := monthStart.AddDate(0, 1, 0)
		for week := weekStart(monthStart); week.Before(nextMonth); week = week.AddDate(0, 0, 7) {
			for day := 0; day < 7; day++ {
				current := week.AddDate(0, 0, day)
				cell := cellBlank
				if !current.Before(monthStart) && current.Before(nextMonth) {
					cell = cellMuted
					if marks[current.Format("2006-01-02")] {
						cell = cellMarked
					}
				}
				grid[day] = append(grid[day], cell)
			}
		}
	}
	return grid
}

func habitMonthHeader(starts []int, labels []string, offsetX, endX int) string {
	header := make([]rune, max(0, endX-offsetX))
	for i := range header {
		header[i] = ' '
	}
	for index, start := range starts {
		if start < offsetX || start >= endX {
			continue
		}
		for i, r := range labels[index] {
			pos := start - offsetX + i
			if pos >= 0 && pos < len(header) {
				header[pos] = r
			}
		}
	}
	return "  " + mutedStyle.Render(string(header))
}

func renderHabitRows(grid [7][]byte, offsetX, endX int) []string {
	labels := []string{"L", "M", "X", "J", "V", "S", "D"}
	lines := make([]string, 7)
	for day := 0; day < 7; day++ {
		var cells strings.Builder
		for col := offsetX; col < endX && col < len(grid[day]); col++ {
			switch grid[day][col] {
			case cellMarked:
				cells.WriteString(lipgloss.NewStyle().Foreground(colorLeaf).Render("●"))
			case cellMuted:
				cells.WriteString(mutedStyle.Render("·"))
			default:
				cells.WriteString(" ")
			}
		}
		lines[day] = mutedStyle.Render(labels[day]) + " " + cells.String()
	}
	return lines
}
