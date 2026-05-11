package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"sword-and-board/internal/adapters/tui/styles"
	"sword-and-board/internal/domain"
)

// StatsKeyMap defines key bindings for the stats view
type StatsKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Use     key.Binding
	Restore key.Binding
	Rest    key.Binding
	Edit    key.Binding
	Enter   key.Binding
	Escape  key.Binding
}

var DefaultStatsKeyMap = StatsKeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Use:     key.NewBinding(key.WithKeys("-")),
	Restore: key.NewBinding(key.WithKeys("+", "=")),
	Rest:    key.NewBinding(key.WithKeys("r")),
	Edit:    key.NewBinding(key.WithKeys("e")),
	Enter:   key.NewBinding(key.WithKeys("enter")),
	Escape:  key.NewBinding(key.WithKeys("esc")),
}

// StatsModel handles the stats view
type StatsModel struct {
	Character     *domain.Character
	selectedIndex int
	editing       bool
	editInput     textinput.Model
	keyMap        StatsKeyMap
}

// NewStatsModel creates a new stats view model
func NewStatsModel(Character *domain.Character) *StatsModel {
	input := textinput.New()
	input.CharLimit = 5
	return &StatsModel{
		Character: Character,
		keyMap:    DefaultStatsKeyMap,
		editInput: input,
	}
}

// IsInputMode returns true when the view is capturing text input
func (m *StatsModel) IsInputMode() bool {
	return m.editing
}

// Init implements tea.Model
func (m *StatsModel) Init() tea.Cmd {
	return nil
}

// numEditableStats is the count of numeric stats that can be modified with +/-
// (Poise, Poise Points, Damage Taken)
const numEditableStats = 3

// totalSelectableItems returns the total number of items the cursor can land on
func (m *StatsModel) totalSelectableItems() int {
	return numEditableStats + len(m.Character.Resources)
}

// Update implements tea.Model
func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.updateEditMode(msg)
		}

		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.selectedIndex < m.totalSelectableItems()-1 {
				m.selectedIndex++
			}
		case key.Matches(msg, m.keyMap.Edit):
			if m.selectedIndex < numEditableStats {
				return m.enterEditMode()
			}
		case key.Matches(msg, m.keyMap.Use):
			updated := m.applyDelta(-1)
			if updated {
				return m, func() tea.Msg {
					return CharacterUpdatedMsg{Character: m.Character}
				}
			}
		case key.Matches(msg, m.keyMap.Restore):
			updated := m.applyDelta(1)
			if updated {
				return m, func() tea.Msg {
					return CharacterUpdatedMsg{Character: m.Character}
				}
			}
		case key.Matches(msg, m.keyMap.Rest):
			m.Character.Rest()
			return m, func() tea.Msg {
				return StatusMsg{Message: "Rested at bonfire. All resources restored.", IsError: false}
			}
		}
	}

	return m, nil
}

func (m *StatsModel) enterEditMode() (tea.Model, tea.Cmd) {
	var current int
	switch m.selectedIndex {
	case 0:
		current = m.Character.Stats.Poise
	case 1:
		current = m.Character.Stats.PoisePoints
	case 2:
		current = m.Character.Stats.DamageTaken
	}
	m.editInput.SetValue(strconv.Itoa(current))
	m.editInput.Focus()
	m.editing = true
	return m, textinput.Blink
}

func (m *StatsModel) updateEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Escape):
		m.editing = false
		return m, nil
	case key.Matches(msg, m.keyMap.Enter):
		val, err := strconv.Atoi(m.editInput.Value())
		if err != nil {
			return m, func() tea.Msg {
				return StatusMsg{Message: "Invalid number", IsError: true}
			}
		}
		switch m.selectedIndex {
		case 0:
			m.Character.Stats.Poise = val
		case 1:
			m.Character.Stats.PoisePoints = val
		case 2:
			if val < 0 {
				val = 0
			}
			m.Character.Stats.DamageTaken = val
		}
		m.editing = false
		return m, func() tea.Msg {
			return CharacterUpdatedMsg{Character: m.Character}
		}
	}

	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

// applyDelta modifies the currently selected stat or resource by delta
func (m *StatsModel) applyDelta(delta int) bool {
	if m.selectedIndex < numEditableStats {
		// Editing a numeric stat
		switch m.selectedIndex {
		case 0: // Poise
			m.Character.Stats.Poise += delta
		case 1: // Poise Points
			m.Character.Stats.PoisePoints += delta
		case 2: // Damage Taken — + heals (decrease), - hurts (increase) to match Poise Points semantics
			m.Character.Stats.DamageTaken -= delta
			if m.Character.Stats.DamageTaken < 0 {
				m.Character.Stats.DamageTaken = 0
			}
		}
		return true
	}

	// Editing a resource
	resIdx := m.selectedIndex - numEditableStats
	if resIdx < len(m.Character.Resources) {
		if delta < 0 {
			return m.Character.Resources[resIdx].Use()
		}
		return m.Character.Resources[resIdx].Restore()
	}
	return false
}

// wideLabel is a label style wide enough for "Parry & Repost"
var wideLabel = styles.LabelStyle.Width(16)

// View implements tea.Model
func (m *StatsModel) View() string {
	var b strings.Builder

	// Combat section — editable stats
	b.WriteString(styles.SectionStyle.Render("COMBAT") + "\n\n")

	editableStats := []struct {
		label string
		value string
	}{
		{"POISE", fmt.Sprintf("%d (%s)", m.Character.Stats.Poise, m.Character.Stats.PoiseDie)},
		{"Poise Points", fmt.Sprintf("%d", m.Character.Stats.PoisePoints)},
		{"Damage", fmt.Sprintf("%d", m.Character.Stats.DamageTaken)},
	}

	for i, stat := range editableStats {
		cursor := "  "
		labelStyle := wideLabel
		if i == m.selectedIndex {
			cursor = "> "
			labelStyle = styles.SelectedStyle.Width(16)
		}
		label := labelStyle.Render(stat.label)
		if m.editing && i == m.selectedIndex {
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, m.editInput.View()))
		} else {
			value := styles.ValueStyle.Render(stat.value)
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, value))
		}
	}

	// HP bar beneath combat stats
	maxHP := m.Character.Stats.EffectiveMaxHP()
	hpBar := styles.RenderHPBar(m.Character.Stats.DamageTaken, maxHP)
	remaining := maxHP - m.Character.Stats.DamageTaken
	if remaining < 0 {
		remaining = 0
	}
	b.WriteString(fmt.Sprintf("\n  %s %s  %s\n",
		wideLabel.Render("HP"),
		hpBar,
		styles.DimmedStyle.Render(fmt.Sprintf("%d/%d", remaining, maxHP)),
	))

	if m.Character.Stats.DamageNote != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			wideLabel.Render(""),
			styles.DimmedStyle.Render(m.Character.Stats.DamageNote),
		))
	}

	// Abilities section — read-only
	b.WriteString("\n")
	b.WriteString(styles.SectionStyle.Render("ABILITIES") + "\n\n")

	readOnlyStats := []struct {
		label string
		value string
	}{
		{"Special", m.Character.Stats.Special},
		{"Parry & Repost", m.Character.Stats.ParryAndRepost},
	}

	for _, stat := range readOnlyStats {
		if stat.value == "" {
			continue
		}
		label := wideLabel.Render(stat.label)
		value := styles.ValueStyle.Render(stat.value)
		b.WriteString(fmt.Sprintf("  %s %s\n", label, value))
	}

	// Resources section
	b.WriteString("\n")
	b.WriteString(styles.SectionStyle.Render("RESOURCES") + "\n\n")

	if len(m.Character.Resources) == 0 {
		b.WriteString(styles.EmptyStateStyle.Render("No resources.") + "\n")
	}

	for i, res := range m.Character.Resources {
		cursor := "  "
		nameStyle := styles.NormalStyle
		if i+numEditableStats == m.selectedIndex {
			cursor = "> "
			nameStyle = styles.SelectedStyle
		}

		dots := styles.RenderResourceDots(res.Remaining, res.Total)
		name := nameStyle.Render(fmt.Sprintf("%-16s", res.Name))
		count := styles.DimmedStyle.Render(fmt.Sprintf("(%d/%d)", res.Remaining, res.Total))

		b.WriteString(fmt.Sprintf("%s%s  %s %s\n", cursor, name, dots, count))
	}

	// Help for this view
	b.WriteString("\n")
	help := []string{
		styles.KeyStyle.Render("[j/k]") + " select",
		styles.KeyStyle.Render("[e]") + " edit",
		styles.KeyStyle.Render("[+/-]") + " use/restore",
		styles.KeyStyle.Render("[r]") + " rest",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}
