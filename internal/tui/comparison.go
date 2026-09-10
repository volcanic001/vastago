package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tslc "github.com/NimbleMarkets/ntcharts/v2/linechart/timeserieslinechart"

	"github.com/volcanic001/vastago/internal/store"
)

const comparisonChartHeight = 8

const previousComparisonDataSet = "previous"

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

// comparisonChart renders both period series with ntcharts' Braille renderer.
// The chart is rebuilt from the current terminal width on every view, so its
// axes, data scaling, and label placement always match the available space.
func comparisonChart(width int, period store.Period, series comparisonSeries) []string {
	if len(series.labels) == 0 || width < 12 {
		return nil
	}

	maximum := comparisonMaximum(series)
	chartWidth := width
	chart := tslc.New(chartWidth, comparisonChartHeight,
		tslc.WithTimeRange(period.Start, period.End),
		tslc.WithYRange(0, maximum.Seconds()),
		tslc.WithXYSteps(1, 2),
		tslc.WithXLabelFormatter(func(int, float64) string { return "" }),
		tslc.WithYLabelFormatter(comparisonDurationLabel),
		tslc.WithAxesStyles(mutedStyle, mutedStyle),
		tslc.WithStyle(titleStyle),
		tslc.WithDataSetStyle(previousComparisonDataSet, mutedStyle),
	)

	for index, point := range comparisonPoints(period, series) {
		valueIndex := min(index, len(series.current)-1)
		chart.Push(tslc.TimePoint{Time: point, Value: series.current[valueIndex].Seconds()})
		chart.PushDataSet(previousComparisonDataSet, tslc.TimePoint{Time: point, Value: series.previous[valueIndex].Seconds()})
	}
	chart.DrawBrailleAll()

	lines := strings.Split(chart.View(), "\n")
	if len(lines) > 0 {
		// ntcharts reserves the last row for X labels. Reuse it so the labels
		// align with the Braille grid and stay evenly distributed at any width.
		lines[len(lines)-1] = comparisonLabels(chartWidth, chart.Origin().X, chart.GraphWidth(), series.labels)
	}
	return append([]string{comparisonHeader(width)}, lines...)
}

func comparisonMaximum(series comparisonSeries) time.Duration {
	maximum := time.Duration(0)
	for _, values := range [][]time.Duration{series.current, series.previous} {
		for _, value := range values {
			if value > maximum {
				maximum = value
			}
		}
	}
	if maximum <= 0 {
		return time.Second
	}
	return maximum
}

func comparisonPoints(period store.Period, series comparisonSeries) []time.Time {
	if len(series.labels) == 1 {
		// A day has one aggregate value. Duplicate it at both ends only for
		// rendering, producing a readable horizontal Braille line.
		return []time.Time{period.Start, period.End}
	}
	points := make([]time.Time, len(series.labels))
	span := period.End.Sub(period.Start)
	for index := range points {
		points[index] = period.Start.Add(time.Duration(index) * span / time.Duration(len(points)-1))
	}
	return points
}

func comparisonDurationLabel(_ int, seconds float64) string {
	return store.FormatDuration(time.Duration(math.Round(seconds)) * time.Second)
}

func comparisonHeader(width int) string {
	if width < 24 {
		return titleStyle.Render("── actual") + mutedStyle.Render(" ·· anterior")
	}
	return mutedStyle.Render("COMPARATIVA · ") + titleStyle.Render("── actual") + mutedStyle.Render("   ·· anterior")
}

func comparisonLabels(width, origin, graphWidth int, labels []string) string {
	if width <= 0 || graphWidth <= 0 || len(labels) == 0 {
		return ""
	}
	line := make([]rune, width)
	for index := range line {
		line[index] = ' '
	}
	start := origin + 1
	end := min(width, start+graphWidth)
	for index, label := range labels {
		position := start
		if len(labels) > 1 {
			position += int(math.Round(float64(index) * float64(graphWidth-1) / float64(len(labels)-1)))
		}
		labelRunes := []rune(label)
		position -= len(labelRunes) / 2
		position = max(start, min(end-len(labelRunes), position))
		for offset, char := range labelRunes {
			if position+offset >= start && position+offset < end {
				line[position+offset] = char
			}
		}
	}
	return mutedStyle.Render(string(line))
}
