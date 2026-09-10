package tui

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// screenLayout gives every screen the same left edge and bottom-aligned footer.
func screenLayout(header, body, footer string, width, height int) string {
	fit := func(content string) []string {
		lines := strings.Split(content, "\n")
		for i := range lines {
			lines[i] = ansi.Truncate(lines[i], width, "")
		}
		return lines
	}
	top, middle, bottom := fit(header), fit(body), fit(footer)
	if height <= 0 {
		height = 24
	}
	// Reserve controls first when the terminal has very little vertical space.
	if len(bottom) > height {
		bottom = bottom[len(bottom)-height:]
	}
	if len(top) > height-len(bottom) {
		top = top[:height-len(bottom)]
	}
	available := height - len(top) - len(bottom)
	if len(middle) > available {
		middle = middle[:available]
	}
	for len(middle) < available {
		middle = append(middle, "")
	}
	lines := append(top, middle...)
	lines = append(lines, bottom...)
	return strings.Join(lines, "\n")
}
