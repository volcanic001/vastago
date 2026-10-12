package tui

import (
	"image/color"
	"testing"
	"time"

	"github.com/volcanic001/vastago/internal/store"
)

// withDefaultTheme resets the theme after the test so changes made via the
// package-level theme state don't leak into unrelated tests.
func withDefaultTheme(t *testing.T) {
	t.Helper()
	setTheme(0)
	t.Cleanup(func() { setTheme(0) })
}

func TestThemesCatalog(t *testing.T) {
	want := []string{"default", "green1999", "amber", "warmamber"}
	if len(themes) != len(want) {
		t.Fatalf("themes = %d entries, want %d", len(themes), len(want))
	}
	for i, name := range want {
		if themes[i].name != name {
			t.Fatalf("themes[%d].name = %q, want %q", i, themes[i].name, name)
		}
	}
}

func TestSetThemeAppliesColorsAndWraps(t *testing.T) {
	withDefaultTheme(t)

	setTheme(1)
	if activeTheme != 1 || colorLeaf != color.Color(themes[1].leaf) {
		t.Fatalf("setTheme(1) did not apply green1999 colors: active=%d leaf=%v", activeTheme, colorLeaf)
	}
	if progressBase != toRGBA(themes[1].leaf) {
		t.Fatalf("progressBase did not follow theme: %+v", progressBase)
	}

	setTheme(len(themes))
	if activeTheme != 0 {
		t.Fatalf("setTheme did not wrap forward: active=%d", activeTheme)
	}
	setTheme(-1)
	if activeTheme != len(themes)-1 {
		t.Fatalf("setTheme did not wrap backward: active=%d", activeTheme)
	}
}

func TestCycleThemeAdvancesAndWraps(t *testing.T) {
	withDefaultTheme(t)

	for i := 1; i < len(themes); i++ {
		name := cycleTheme()
		if name != themes[i].name || activeTheme != i {
			t.Fatalf("cycleTheme() = %q (active=%d), want %q (active=%d)", name, activeTheme, themes[i].name, i)
		}
	}
	name := cycleTheme()
	if name != themes[0].name || activeTheme != 0 {
		t.Fatalf("cycleTheme() did not wrap to default: %q (active=%d)", name, activeTheme)
	}
}

func TestThemeKeyOnlyWorksOnMetricsScreen(t *testing.T) {
	withDefaultTheme(t)

	now := time.Date(2026, 6, 15, 9, 0, 0, 0, time.Local)
	m := Model{db: &store.Database{}, now: now, screen: metricsScreen, width: 80, height: 24}
	updated, _ := m.handleKey(keyMessage('c', "c"))
	m = updated.(Model)
	if activeTheme != 1 || m.message != "tema: "+themes[1].name {
		t.Fatalf("c on metrics did not cycle theme: active=%d message=%q", activeTheme, m.message)
	}

	m.screen = sessionsScreen
	updated, _ = m.handleKey(keyMessage('c', "c"))
	m = updated.(Model)
	if activeTheme != 1 {
		t.Fatalf("c outside metrics changed theme: active=%d", activeTheme)
	}
}

func TestViewUsesActiveThemeBackground(t *testing.T) {
	withDefaultTheme(t)

	m := Model{db: &store.Database{}, now: time.Now(), width: 40, height: 20, screen: sessionsScreen}
	if got := m.View().BackgroundColor; got != colorBackground {
		t.Fatalf("View().BackgroundColor = %v, want %v", got, colorBackground)
	}
	setTheme(2)
	if got := m.View().BackgroundColor; got != colorBackground {
		t.Fatalf("View().BackgroundColor after theme change = %v, want %v", got, colorBackground)
	}
}
