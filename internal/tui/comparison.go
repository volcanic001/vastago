package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/volcanic001/vastago/internal/store"
)

type comparisonSeries struct {
	labels   []string
	current  []time.Duration
	previous []time.Duration
}

func buildComparisonSeries(period store.Period, current, previous []store.DayActivity) comparisonSeries {
	switch period.Kind {
	case store.Week:
		return comparisonSeries{
			labels:   []string{"L", "M", "X", "J", "V", "S", "D"},
			current:  activityDurations(current, 7),
			previous: activityDurations(previous, 7),
		}
	case store.Month:
		count := max((len(current)+6)/7, (len(previous)+6)/7)
		labels := make([]string, count)
		for index := range labels {
			labels[index] = fmt.Sprintf("S%d", index+1)
		}
		return comparisonSeries{labels: labels, current: groupedActivity(current, 7, count), previous: groupedActivity(previous, 7, count)}
	case store.Year:
		labels := []string{"E", "F", "M", "A", "M", "J", "J", "A", "S", "O", "N", "D"}
		return comparisonSeries{labels: labels, current: monthlyActivity(current), previous: monthlyActivity(previous)}
	default:
		return comparisonSeries{labels: []string{"D"}, current: activityDurations(current, 1), previous: activityDurations(previous, 1)}
	}
}

func activityDurations(activity []store.DayActivity, count int) []time.Duration {
	result := make([]time.Duration, count)
	for index, day := range activity {
		if index >= len(result) {
			break
		}
		result[index] = day.Duration
	}
	return result
}

func groupedActivity(activity []store.DayActivity, groupSize, count int) []time.Duration {
	result := make([]time.Duration, count)
	for index, day := range activity {
		group := index / groupSize
		if group >= len(result) {
			break
		}
		result[group] += day.Duration
	}
	return result
}

func monthlyActivity(activity []store.DayActivity) []time.Duration {
	result := make([]time.Duration, 12)
	for _, day := range activity {
		result[int(day.Date.Month())-1] += day.Duration
	}
	return result
}

// comparisonChart renders the selected period as a solid line and its prior
// equivalent as a muted dotted line. It scales both lines with one shared Y axis.
func comparisonChart(width int, series comparisonSeries) []string {
	if len(series.labels) == 0 || width < 12 {
		return nil
	}
	maximum := time.Duration(0)
	for _, values := range [][]time.Duration{series.current, series.previous} {
		for _, value := range values {
			if value > maximum {
				maximum = value
			}
		}
	}
	axisWidth := max(2, lipgloss.Width(store.FormatDuration(maximum)))
	step := 2
	plotWidth := len(series.labels)*step - 1
	if axisWidth+1+plotWidth > width {
		step = 1
		plotWidth = len(series.labels)
	}
	if axisWidth+1+plotWidth > width {
		return nil
	}

	header := "COMPARATIVA · ── actual   ·· anterior"
	if width < 36 {
		header = "── actual ·· anterior"
	}
	const height = 5
	grid := make([][]chartCell, height)
	for row := range grid {
		grid[row] = make([]chartCell, plotWidth)
	}
	drawComparisonLine(grid, series.previous, maximum, step, chartPreviousLine, chartPreviousPoint)
	drawComparisonLine(grid, series.current, maximum, step, chartCurrentLine, chartCurrentPoint)

	lines := []string{mutedStyle.Render(trimToWidth(header, width))}
	for row := range grid {
		axis := ""
		switch row {
		case 0:
			axis = store.FormatDuration(maximum)
		case height / 2:
			axis = store.FormatDuration(maximum / 2)
		case height - 1:
			axis = "0s"
		}
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("%*s", axisWidth, axis))+" "+renderChartRow(grid[row]))
	}
	labels := make([]rune, plotWidth)
	for index := range labels {
		labels[index] = ' '
	}
	for index, label := range series.labels {
		position := index * step
		for offset, char := range []rune(label) {
			if position+offset < len(labels) {
				labels[position+offset] = char
			}
		}
	}
	lines = append(lines, strings.Repeat(" ", axisWidth+1)+mutedStyle.Render(string(labels)))
	return lines
}

type chartCell uint8

const (
	chartEmpty chartCell = iota
	chartPreviousLine
	chartPreviousPoint
	chartCurrentLine
	chartCurrentPoint
)

func drawComparisonLine(grid [][]chartCell, values []time.Duration, maximum time.Duration, step int, line, point chartCell) {
	if len(values) == 0 {
		return
	}
	for index, value := range values {
		x := index * step
		y := comparisonY(value, maximum, len(grid))
		setChartCell(grid, x, y, point)
		if index == 0 {
			continue
		}
		previousY := comparisonY(values[index-1], maximum, len(grid))
		for between := 1; between < step; between++ {
			setChartCell(grid, x-step+between, (previousY+y)/2, line)
		}
	}
}

func comparisonY(value, maximum time.Duration, height int) int {
	if maximum <= 0 || height <= 1 {
		return max(0, height-1)
	}
	position := int(math.Round(float64(maximum-value) / float64(maximum) * float64(height-1)))
	return max(0, min(height-1, position))
}

func setChartCell(grid [][]chartCell, x, y int, kind chartCell) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[y]) || grid[y][x] > kind {
		return
	}
	grid[y][x] = kind
}

func renderChartRow(row []chartCell) string {
	var output strings.Builder
	for _, cell := range row {
		switch cell {
		case chartCurrentPoint:
			output.WriteString(titleStyle.Render("●"))
		case chartCurrentLine:
			output.WriteString(titleStyle.Render("─"))
		case chartPreviousPoint, chartPreviousLine:
			output.WriteString(mutedStyle.Render("·"))
		default:
			output.WriteByte(' ')
		}
	}
	return output.String()
}
