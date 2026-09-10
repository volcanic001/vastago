package tui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	progressStart = color.RGBA{R: 34, G: 197, B: 94, A: 255}
	progressEnd   = color.RGBA{R: 163, G: 230, B: 53, A: 255}
	progressEmpty = lipgloss.NewStyle().Foreground(lipgloss.Color("#34413A"))
)

// focusProgressBar draws a compact btop-like meter. Its filled cells progress
// from green to lime and its empty cells remain visible on dark terminals.
func focusProgressBar(width int, fraction float64) string {
	if width < 3 {
		return ""
	}
	fraction = math.Max(0, math.Min(1, fraction))
	cells := width - 2
	filled := int(math.Round(fraction * float64(cells)))
	if fraction > 0 && filled == 0 {
		filled = 1
	}

	var bar strings.Builder
	bar.WriteString(mutedStyle.Render("▏"))
	for i := 0; i < cells; i++ {
		if i >= filled {
			bar.WriteString(progressEmpty.Render("░"))
			continue
		}
		bar.WriteString(lipgloss.NewStyle().Foreground(progressColor(i, filled)).Render("█"))
	}
	bar.WriteString(mutedStyle.Render("▕"))
	return bar.String()
}

func progressColor(index, total int) color.RGBA {
	if total <= 1 {
		return progressEnd
	}
	ratio := float64(index) / float64(total-1)
	return color.RGBA{
		R: uint8(math.Round(float64(progressStart.R) + (float64(progressEnd.R)-float64(progressStart.R))*ratio)),
		G: uint8(math.Round(float64(progressStart.G) + (float64(progressEnd.G)-float64(progressStart.G))*ratio)),
		B: uint8(math.Round(float64(progressStart.B) + (float64(progressEnd.B)-float64(progressStart.B))*ratio)),
		A: 255,
	}
}

func progressPercent(fraction float64) string {
	return fmt.Sprintf("%.0f%%", math.Max(0, math.Min(1, fraction))*100)
}
