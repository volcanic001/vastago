package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestFocusProgressBarUsesFullWidthAndSegments(t *testing.T) {
	bar := focusProgressBar(18, 0.5)
	plain := ansi.Strip(bar)
	if got := len([]rune(plain)); got != 18 {
		t.Fatalf("bar width = %d, want 18: %q", got, plain)
	}
	if !strings.Contains(plain, "█") || !strings.Contains(plain, "░") {
		t.Fatalf("bar is not segmented: %q", plain)
	}
	dark, bright := progressColor(0, 5), progressColor(4, 5)
	if dark.R >= bright.R || dark.G >= bright.G || dark.B >= bright.B {
		t.Fatalf("bar did not brighten: %#v to %#v", dark, bright)
	}
	if progressPercent(1.2) != "100%" || progressPercent(-1) != "0%" {
		t.Fatal("percentage must remain bounded")
	}
}
