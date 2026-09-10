package tui

import (
	"image/color"
	"math"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

var heatmapColors = []color.Color{
	lipgloss.Color("#586257"),
	lipgloss.Color("#668743"),
	lipgloss.Color("#8FAE4F"),
	lipgloss.Color("#D4B65A"),
}

func focusHeatmap(width int, activity []store.DayActivity) []string {
	if len(activity) == 0 || width < 4 {
		return nil
	}
	byDay := make(map[time.Time]time.Duration, len(activity))
	maxDuration := time.Duration(0)
	for _, item := range activity {
		day := store.StartOfDay(item.Date)
		byDay[day] = item.Duration
		if item.Duration > maxDuration {
			maxDuration = item.Duration
		}
	}
	first := activity[0].Date
	first = store.StartOfDay(first).AddDate(0, 0, -(int(first.Weekday())+6)%7)
	last := store.StartOfDay(activity[len(activity)-1].Date)
	last = last.AddDate(0, 0, 6-(int(last.Weekday())+6)%7)
	weeks := int(last.Sub(first).Hours()/24)/7 + 1
	columns := max(1, width-3)
	if weeks > columns {
		first = first.AddDate(0, 0, 7*(weeks-columns))
		weeks = columns
	}
	lines := []string{mutedStyle.Render(trimToWidth("ACTIVIDAD · menos ··· más", width))}
	labels := []string{"L", "M", "X", "J", "V", "S", "D"}
	for row, label := range labels {
		line := mutedStyle.Render(label + " ")
		for week := 0; week < weeks; week++ {
			day := first.AddDate(0, 0, week*7+row)
			line += heatmapDot(byDay[day], maxDuration)
		}
		lines = append(lines, line)
	}
	return lines
}

func heatmapDot(duration, maximum time.Duration) string {
	if duration <= 0 || maximum <= 0 {
		return mutedStyle.Render("·")
	}
	level := int(math.Ceil(4 * float64(duration) / float64(maximum)))
	level = max(1, min(level, len(heatmapColors)))
	return lipgloss.NewStyle().Foreground(heatmapColors[level-1]).Render("·")
}
