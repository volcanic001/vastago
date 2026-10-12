package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

const (
	wideBreakpoint    = 88
	compactBreakpoint = 52
)

// theme holds every color role the TUI draws with. Switching themes
// reassigns the package-level color and style vars below instead of
// threading a theme value through every render function.
type theme struct {
	name string

	background color.Color // full-screen background
	surface    color.Color // dark text drawn on top of a gold/selection bar

	leaf     color.Color // brand accent: titles, list bullets
	gold     color.Color // selection bar, primary emphasis
	amber    color.Color // date stamps, secondary accent
	cream    color.Color // main bold text
	muted    color.Color // secondary/muted text
	shortcut color.Color // footer key hints
	danger   color.Color // destructive actions (delete, errors)

	progressEmpty color.Color // empty track of the focus progress bar
}

var themes = []theme{
	{
		name:          "default",
		background:    lipgloss.Color("#18201B"),
		surface:       lipgloss.Color("#18201B"),
		leaf:          lipgloss.Color("#A7C957"),
		gold:          lipgloss.Color("#E9C46A"),
		amber:         lipgloss.Color("#A88752"),
		cream:         lipgloss.Color("#F4F1DE"),
		muted:         lipgloss.Color("#8A9185"),
		shortcut:      lipgloss.Color("#B6AD8A"),
		danger:        lipgloss.Color("#E07A5F"),
		progressEmpty: lipgloss.Color("#34413A"),
	},
	{
		// "Terminal Green 1999" (theebest.com, also in
		// mbadolato/iTerm2-Color-Schemes): black-glass phosphor CRT green.
		// gold reuses the same green as leaf — a separate gold/amber hue
		// doesn't belong in a single-phosphor terminal aesthetic.
		name:          "green1999",
		background:    lipgloss.Color("#03120a"),
		surface:       lipgloss.Color("#03120a"),
		leaf:          lipgloss.Color("#22c55e"),
		gold:          lipgloss.Color("#22c55e"),
		amber:         lipgloss.Color("#61ffa0"),
		cream:         lipgloss.Color("#d8ffe6"),
		muted:         lipgloss.Color("#8fd6a9"),
		shortcut:      lipgloss.Color("#97dcfd"),
		danger:        lipgloss.Color("#ff5c7a"),
		progressEmpty: lipgloss.Color("#32463a"),
	},
	{
		// github.com/zigmoo/omarchy-amber-console-theme: a single amber
		// emitter at different intensities, so several roles intentionally
		// share the same hex instead of borrowing another hue.
		name:          "amber",
		background:    lipgloss.Color("#100600"),
		surface:       lipgloss.Color("#1e0c00"),
		leaf:          lipgloss.Color("#ffa86d"),
		gold:          lipgloss.Color("#ff6b08"),
		amber:         lipgloss.Color("#dd5800"),
		cream:         lipgloss.Color("#ffa86d"),
		muted:         lipgloss.Color("#dd5800"),
		shortcut:      lipgloss.Color("#dd5800"),
		danger:        lipgloss.Color("#ff6b08"),
		progressEmpty: lipgloss.Color("#1b0c02"),
	},
	{
		// "Ambience: Amber" from github.com/Axel-Ollivier/ambience — a
		// warmer, more cinematic amber with real hue variety (unlike the
		// monochrome amber theme above), including a genuine red for danger.
		name:          "warmamber",
		background:    lipgloss.Color("#16110D"),
		surface:       lipgloss.Color("#120D08"),
		leaf:          lipgloss.Color("#FFB454"),
		gold:          lipgloss.Color("#FFB454"),
		amber:         lipgloss.Color("#C98A52"),
		cream:         lipgloss.Color("#E8D5BC"),
		muted:         lipgloss.Color("#7E6549"),
		shortcut:      lipgloss.Color("#CC785C"),
		danger:        lipgloss.Color("#E8743B"),
		progressEmpty: lipgloss.Color("#221A11"),
	},
}

var activeTheme int

var (
	colorBackground color.Color
	colorSurface    color.Color
	colorLeaf       color.Color
	colorGold       color.Color
	colorAmber      color.Color
	colorCream      color.Color
	colorMuted      color.Color
	colorShortcut   color.Color
	colorDanger     color.Color

	titleStyle       lipgloss.Style
	mutedStyle       lipgloss.Style
	guideStyle       lipgloss.Style
	valueStyle       lipgloss.Style
	dateStyle        lipgloss.Style
	goldStyle        lipgloss.Style
	listMarkerStyle  lipgloss.Style
	errorStyle       lipgloss.Style
	shortcutKeyStyle lipgloss.Style
	selectionStyle   lipgloss.Style
)

func init() {
	setTheme(0)
}

// setTheme rebuilds every color and derived style from themes[index]. All
// renderers read the package vars at render time, so a change here is
// visible on the very next frame without touching call sites elsewhere.
func setTheme(index int) {
	activeTheme = ((index % len(themes)) + len(themes)) % len(themes)
	t := themes[activeTheme]

	colorBackground, colorSurface = t.background, t.surface
	colorLeaf, colorGold, colorAmber = t.leaf, t.gold, t.amber
	colorCream, colorMuted, colorShortcut, colorDanger = t.cream, t.muted, t.shortcut, t.danger

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorLeaf)
	mutedStyle = lipgloss.NewStyle().Foreground(colorMuted)
	guideStyle = lipgloss.NewStyle().Foreground(colorMuted).Faint(true)
	valueStyle = lipgloss.NewStyle().Bold(true).Foreground(colorCream)
	dateStyle = lipgloss.NewStyle().Foreground(colorAmber)
	goldStyle = lipgloss.NewStyle().Bold(true).Foreground(colorGold)
	listMarkerStyle = lipgloss.NewStyle().Bold(true).Foreground(colorLeaf)
	errorStyle = lipgloss.NewStyle().Foreground(colorDanger)
	shortcutKeyStyle = lipgloss.NewStyle().Foreground(colorShortcut)
	selectionStyle = lipgloss.NewStyle().Bold(true).Foreground(colorSurface).Background(colorGold)

	progressBase = toRGBA(colorLeaf)
	progressEmpty = lipgloss.NewStyle().Foreground(t.progressEmpty)
}

// cycleTheme advances to the next theme and returns its name.
func cycleTheme() string {
	setTheme(activeTheme + 1)
	return themes[activeTheme].name
}

func toRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}
}
