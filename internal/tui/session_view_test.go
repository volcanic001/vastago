package tui

import (
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/volcanic001/vastago/internal/store"
)

func TestSessionListAndDetailUseLocalTime(t *testing.T) {
	previous := time.Local
	time.Local = time.FixedZone("user", -6*60*60)
	t.Cleanup(func() { time.Local = previous })

	start := time.Date(2026, time.September, 10, 17, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	model := Model{
		db:     &store.Database{Entries: []store.Entry{{ID: "one", Task: "Trabajo", Note: "Revisar diseño", Start: start, End: &end}}},
		now:    end,
		width:  96,
		height: 24,
		screen: sessionsScreen,
	}
	list := ansi.Strip(model.sessionView(model.width))
	if !strings.Contains(list, "10/09 11:00–12:00 · Trabajo") || !strings.Contains(list, "1h 00m") {
		t.Fatalf("session list did not use local range: %q", list)
	}

	model = pressSession(model, 'i', "i")
	detail := ansi.Strip(model.sessionView(model.width))
	for _, want := range []string{"DETALLE DE SESION", "Inicio 10/09/2026 11:00", "Fin 10/09/2026 12:00", "Revisar diseño"} {
		if !strings.Contains(detail, want) {
			t.Fatalf("detail missing %q: %q", want, detail)
		}
	}
	model = pressSession(model, 'i', "i")
	if model.sessionDetailID != "" {
		t.Fatal("detail did not close")
	}
}

func TestSessionListShowsEndDateWhenDifferent(t *testing.T) {
	start := time.Date(2026, time.September, 11, 10, 13, 0, 0, time.Local)
	end := time.Date(2026, time.September, 15, 10, 0, 0, 0, time.Local)
	line := sessionListLine(store.Entry{Task: "Legacy", Start: start, End: &end}, end, 80, false)
	if !strings.Contains(ansi.Strip(line), "11/09 10:13–15/09 10:00") {
		t.Fatalf("different end date is hidden: %q", line)
	}
}

func TestSelectedSessionLineHasNoInnerStyleReset(t *testing.T) {
	start := time.Date(2026, time.September, 11, 10, 31, 0, 0, time.Local)
	end := start.Add(time.Hour)
	line := sessionListLine(store.Entry{Task: "Trabajo", Start: start, End: &end}, end, 80, true)
	if strings.Contains(line, "\x1b[") {
		t.Fatalf("selected row contains an inner ANSI style reset: %q", line)
	}
}

func TestSessionListAlignsDurationToRightEdge(t *testing.T) {
	start := time.Date(2026, time.September, 10, 9, 0, 0, 0, time.Local)
	end := start.Add(time.Hour)
	width := 48
	line := sessionListLine(store.Entry{
		Task:  "Una tarea con un nombre deliberadamente largo",
		Start: start,
		End:   &end,
	}, end, width, false)

	duration := "1h 00m"
	if !strings.HasSuffix(line, duration) {
		t.Fatalf("duration is not at the end: %q", line)
	}
	if got := lipgloss.Width(line); got != width {
		t.Fatalf("line width = %d, want %d: %q", got, width, line)
	}
}
