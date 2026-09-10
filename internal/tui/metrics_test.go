package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func metricsFixture() Model {
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.Local)
	end := now.Add(-time.Hour)
	db := &store.Database{Entries: []store.Entry{{ID: "one", Task: "Lectura", Start: end.Add(-2 * time.Hour), End: &end}},
		Todos:  []store.Todo{{CompletedAt: &end}},
		Habits: []store.Habit{{CreatedAt: now.Add(-time.Hour), Completions: []string{"2026-01-15"}}}}
	return Model{db: db, now: now, width: 96, height: 30, screen: metricsScreen}
}

func metricKey(m Model, key rune) Model {
	updated, _ := m.handleKey(keyMessage(key, string(key)))
	return updated.(Model)
}

func TestMetricsPeriodNavigation(t *testing.T) {
	m := metricsFixture()
	for key, want := range map[rune]store.PeriodKind{'d': store.Day, 'w': store.Week, 'm': store.Month, 'y': store.Year} {
		m = metricKey(m, key)
		p, err := m.selectedPeriod()
		if err != nil || p.Kind != want {
			t.Fatalf("%c: %+v %v", key, p, err)
		}
	}
	m = metricKey(m, 'm')
	m = metricKey(m, '[')
	p, _ := m.selectedPeriod()
	if p.Start.Year() != 2025 || p.Start.Month() != time.December {
		t.Fatal(p)
	}
	m = metricKey(m, ']')
	p, _ = m.selectedPeriod()
	if p.Start.Year() != 2026 || p.Start.Month() != time.January {
		t.Fatal(p)
	}
	m = metricKey(m, 't')
	if !m.metricsAnchor.IsZero() {
		t.Fatal("current mode not restored")
	}
	m.now = m.now.AddDate(0, 1, 0)
	p, _ = m.selectedPeriod()
	if p.Start.Month() != time.February {
		t.Fatal("current period did not follow clock")
	}
	m = metricKey(m, '[')
	p, _ = m.selectedPeriod()
	m.now = m.now.AddDate(0, 1, 0)
	later, _ := m.selectedPeriod()
	if !later.Start.Equal(p.Start) {
		t.Fatal("historical selection drifted")
	}
}

func TestMetricsResultsAndFuture(t *testing.T) {
	m := metricKey(metricsFixture(), 'd')
	content := ansi.Strip(m.View().Content)
	for _, want := range []string{"15/01/2026", "Tiempo: 2h 00m", "Sesiones: 1", "Pendientes hechos: 1", "Habitos: 1/1 · 100%", "ACTIVIDAD", "L ", "Lectura", "█"} {
		if !strings.Contains(content, want) {
			t.Fatalf("missing %q: %s", want, content)
		}
	}
	m = metricKey(m, ']')
	content = ansi.Strip(m.View().Content)
	if !strings.Contains(content, "Tiempo: 0s") || !strings.Contains(content, "Periodo futuro") {
		t.Fatal(content)
	}
}

func TestMetricsScrollAndInputIsolation(t *testing.T) {
	m := metricsFixture()
	m.width, m.height = 16, 12
	for i := 0; i < 100; i++ {
		m = metricKey(m, 'j')
	}
	_, limit := m.metricsPage(m.width)
	if m.metricsScroll != limit || limit == 0 {
		t.Fatal("scroll did not stop at bottom")
	}
	m = metricKey(m, 'k')
	if m.metricsScroll != limit-1 {
		t.Fatal("scroll back failed")
	}
	m = metricKey(m, 'y')
	if m.metricsScroll != 0 {
		t.Fatal("period change did not reset scroll")
	}
	before, _ := m.selectedPeriod()
	m.inputMode = true
	for _, key := range "dwmy[]" {
		m = metricKey(m, key)
	}
	after, _ := m.selectedPeriod()
	if string(m.input) != "dwmy[]" || !after.Start.Equal(before.Start) || after.Kind != before.Kind {
		t.Fatal("input changed metrics period")
	}
	m.inputMode = false
	updated, _ := m.handleKey(tea.KeyPressMsg(tea.Key{Code: tea.KeyTab}))
	if updated.(Model).screen != homeScreen {
		t.Fatal("global navigation broken")
	}
}

func TestMetricsReadOnly(t *testing.T) {
	m := metricsFixture()
	entries := len(m.db.Entries)
	m = metricKey(m, 'd')
	m = metricKey(m, '[')
	_ = m.View()
	if len(m.db.Entries) != entries || m.confirmTodo || m.confirmHabit || m.sessionDeleteID != "" {
		t.Fatal("metrics navigation mutated data")
	}
}
