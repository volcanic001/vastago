package tui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestComparisonChartShowsWeeklySeriesAndLegend(t *testing.T) {
	period, err := store.NewPeriod(store.Week, time.Date(2026, time.September, 10, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	current := makeActivity(period.Start, []time.Duration{time.Hour, 2 * time.Hour, 0, 3 * time.Hour, time.Hour, 0, 2 * time.Hour})
	previous := makeActivity(period.Start.AddDate(0, 0, -7), []time.Duration{0, time.Hour, time.Hour, 0, 2 * time.Hour, time.Hour, 0})
	chart := ansi.Strip(strings.Join(comparisonChart(80, period, buildComparisonSeries(period, current, previous)), "\n"))
	for _, want := range []string{"COMPARATIVA", "── actual", "·· anterior", "L", "M", "X", "J", "V", "S", "D", "3h 00m"} {
		if !strings.Contains(chart, want) {
			t.Fatalf("chart missing %q: %q", want, chart)
		}
	}
	baseline := ""
	for _, line := range strings.Split(chart, "\n") {
		if strings.ContainsRune(line, '└') {
			baseline = line
			break
		}
	}
	corner := strings.IndexRune(baseline, '└')
	if baseline == "" || corner < 0 || strings.Trim(baseline[corner+len("└"):], "─ ") != "" {
		t.Fatalf("X axis baseline is not one constant horizontal row: %q", baseline)
	}
}

func TestComparisonSeriesAdaptsToEachPeriod(t *testing.T) {
	anchor := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.Local)
	for _, kind := range []store.PeriodKind{store.Day, store.Week, store.Month, store.Year} {
		period, err := store.NewPeriod(kind, anchor)
		if err != nil {
			t.Fatal(err)
		}
		activity := makeActivity(period.Start, make([]time.Duration, int(period.End.Sub(period.Start).Hours()/24)))
		series := buildComparisonSeries(period, activity, activity)
		if len(series.labels) == 0 || len(series.current) != len(series.labels) || len(series.previous) != len(series.labels) {
			t.Fatalf("%s series malformed: %#v", kind, series)
		}
	}
}

func makeActivity(start time.Time, durations []time.Duration) []store.DayActivity {
	activity := make([]store.DayActivity, len(durations))
	for index, duration := range durations {
		activity[index] = store.DayActivity{Date: start.AddDate(0, 0, index), Duration: duration}
	}
	return activity
}

func TestComparisonChartUsesTerminalWidthAndDistributedLabels(t *testing.T) {
	period, err := store.NewPeriod(store.Week, time.Date(2026, time.September, 10, 12, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	series := buildComparisonSeries(
		period,
		makeActivity(period.Start, []time.Duration{time.Hour, 2 * time.Hour, time.Hour, 3 * time.Hour, 2 * time.Hour, time.Hour, 2 * time.Hour}),
		makeActivity(period.Start.AddDate(0, 0, -7), []time.Duration{2 * time.Hour, time.Hour, 2 * time.Hour, time.Hour, 2 * time.Hour, time.Hour, 2 * time.Hour}),
	)
	for _, width := range []int{70, 90} {
		chart := comparisonChart(width, period, series)
		longest := 0
		for _, line := range chart {
			if got := lipgloss.Width(line); got > width {
				t.Fatalf("width %d: chart overflowed to %d: %q", width, got, line)
			} else if got > longest {
				longest = got
			}
		}
		if longest < width-1 {
			t.Fatalf("width %d: chart only used %d columns", width, longest)
		}

		labels := []rune(ansi.Strip(chart[len(chart)-1]))
		first := strings.IndexRune(string(labels), 'L')
		last := strings.LastIndex(string(labels), "D")
		if first < 0 || last < width-10 {
			t.Fatalf("width %d: weekday labels were not spread across chart: %q", width, string(labels))
		}
	}
}
