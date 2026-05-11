package styles

import "github.com/charmbracelet/lipgloss"

// Colors - Dark Souls inspired palette
var (
	ColorGold      = lipgloss.Color("#D4AF37")
	ColorSilver    = lipgloss.Color("#C0C0C0")
	ColorDark      = lipgloss.Color("#1a1a1a")
	ColorDarkGray  = lipgloss.Color("#333333")
	ColorLightGray = lipgloss.Color("#888888")
	ColorRed       = lipgloss.Color("#8B0000")
	ColorGreen     = lipgloss.Color("#228B22")
	ColorBlue      = lipgloss.Color("#4169E1")
	ColorOrange    = lipgloss.Color("#FF8C00")
)

// Base styles
var (
	// Title style for character name
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold).
			MarginBottom(1)

	// Subtitle for class/path info
	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSilver).
			Italic(true)

	// Section header style
	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorDarkGray).
			MarginTop(1).
			PaddingBottom(0)

	// Label style for stat names
	LabelStyle = lipgloss.NewStyle().
			Foreground(ColorSilver).
			Width(14)

	// Value style for stat values
	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Selected item style
	SelectedStyle = lipgloss.NewStyle().
			Foreground(ColorGold).
			Bold(true)

	// Normal item style
	NormalStyle = lipgloss.NewStyle().
			Foreground(ColorSilver)

	// Dimmed style for notes/secondary info
	DimmedStyle = lipgloss.NewStyle().
			Foreground(ColorLightGray).
			Italic(true)

	// Help bar style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorLightGray).
			MarginTop(1)

	// Key style for help text
	KeyStyle = lipgloss.NewStyle().
			Foreground(ColorGold)

	// Box style for main content
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDarkGray).
			Padding(1, 2)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGold).
			Background(ColorDarkGray).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorLightGray).
				Padding(0, 2)

	// Resource indicators
	FilledDot = lipgloss.NewStyle().
			Foreground(ColorGreen).
			SetString("●")

	EmptyDot = lipgloss.NewStyle().
			Foreground(ColorRed).
			SetString("○")

	// Spell indicators (same as resources but different semantics)
	SpellAvailable = lipgloss.NewStyle().
			Foreground(ColorBlue).
			SetString("●")

	SpellUsed = lipgloss.NewStyle().
			Foreground(ColorDarkGray).
			SetString("○")

	// HP bar styles
	HPFilled = lipgloss.NewStyle().
			Foreground(ColorGreen).
			SetString("█")

	HPEmpty = lipgloss.NewStyle().
		Foreground(ColorRed).
		SetString("░")

	// Tab separator line
	TabSeparatorStyle = lipgloss.NewStyle().
				Foreground(ColorDarkGray)

	// Empty state hint
	EmptyStateStyle = lipgloss.NewStyle().
			Foreground(ColorLightGray).
			Italic(true).
			MarginLeft(2)

	// Item type colors
	WeaponStyle = lipgloss.NewStyle().
			Foreground(ColorOrange)

	EquipmentStyle = lipgloss.NewStyle().
			Foreground(ColorBlue)

	ConsumableStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	// Status messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorGreen)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorRed)
)

// RenderDots renders a series of filled and empty dots for resources/spells
func RenderDots(filled, total int, filledStyle, emptyStyle lipgloss.Style) string {
	result := ""
	for i := 0; i < total; i++ {
		if i > 0 {
			result += " "
		}
		if i < filled {
			result += filledStyle.String()
		} else {
			result += emptyStyle.String()
		}
	}
	return result
}

// RenderResourceDots renders resource dots (filled = remaining)
func RenderResourceDots(remaining, total int) string {
	return RenderDots(remaining, total, FilledDot, EmptyDot)
}

// RenderSpellDots renders spell dots (filled = remaining, empty = used)
func RenderSpellDots(used, total int) string {
	remaining := total - used
	return RenderDots(remaining, total, SpellAvailable, SpellUsed)
}

// RenderHPBar renders a visual HP bar given damage taken and a max HP estimate
func RenderHPBar(damageTaken, maxHP int) string {
	remaining := maxHP - damageTaken
	if remaining < 0 {
		remaining = 0
	}
	result := ""
	for i := 0; i < maxHP; i++ {
		if i < remaining {
			result += HPFilled.String()
		} else {
			result += HPEmpty.String()
		}
	}
	return result
}
