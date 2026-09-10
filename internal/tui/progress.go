package tui

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	progressBase  = color.RGBA{R: 110, G: 204, B: 94, A: 255}
	progressDark  = 0.36
	progressEmpty = lipgloss.NewStyle().Foreground(lipgloss.Color("#34413A"))
)

// focusProgressBar draws a compact btop-like meter. Its filled cells keep one
// green hue and brighten toward their end. Empty cells remain visible on dark terminals.
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
		return scaleProgressColor(1)
	}
	ratio := float64(index) / float64(total-1)
	return scaleProgressColor(progressDark + (1-progressDark)*ratio)
}

func scaleProgressColor(brightness float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Round(float64(progressBase.R) * brightness)),
		G: uint8(math.Round(float64(progressBase.G) * brightness)),
		B: uint8(math.Round(float64(progressBase.B) * brightness)),
		A: 255,
	}
}

func progressPercent(fraction float64) string {
	return fmt.Sprintf("%.0f%%", math.Max(0, math.Min(1, fraction))*100)
}
