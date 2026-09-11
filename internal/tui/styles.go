package tui

import "charm.land/lipgloss/v2"

const (
	wideBreakpoint    = 88
	compactBreakpoint = 52
)

var (
	colorLeaf     = lipgloss.Color("#A7C957")
	colorGold     = lipgloss.Color("#E9C46A")
	colorCream    = lipgloss.Color("#F4F1DE")
	colorMuted    = lipgloss.Color("#8A9185")
	colorShortcut = lipgloss.Color("#B6AD8A")
	colorDanger   = lipgloss.Color("#E07A5F")
	colorSurface  = lipgloss.Color("#18201B")

	titleStyle       = lipgloss.NewStyle().Bold(true).Foreground(colorLeaf)
	mutedStyle       = lipgloss.NewStyle().Foreground(colorMuted)
	valueStyle       = lipgloss.NewStyle().Bold(true).Foreground(colorCream)
	goldStyle        = lipgloss.NewStyle().Bold(true).Foreground(colorGold)
	errorStyle       = lipgloss.NewStyle().Foreground(colorDanger)
	shortcutKeyStyle = lipgloss.NewStyle().Foreground(colorShortcut)
	selectionStyle   = lipgloss.NewStyle().Bold(true).Foreground(colorSurface).Background(colorGold)
)
