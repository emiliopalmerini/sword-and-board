package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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
}

var DefaultStatsKeyMap = StatsKeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Use:     key.NewBinding(key.WithKeys("-")),
	Restore: key.NewBinding(key.WithKeys("+", "=")),
	Rest:    key.NewBinding(key.WithKeys("r")),
}

// StatsModel handles the stats view
type StatsModel struct {
	Character     *domain.Character
	selectedIndex int
	keyMap        StatsKeyMap
}

// NewStatsModel creates a new stats view model
func NewStatsModel(Character *domain.Character) *StatsModel {
	return &StatsModel{
		Character: Character,
		keyMap:    DefaultStatsKeyMap,
	}
}

// Init implements tea.Model
func (m *StatsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.selectedIndex < len(m.Character.Resources)-1 {
				m.selectedIndex++
			}
		case key.Matches(msg, m.keyMap.Use):
			if m.selectedIndex < len(m.Character.Resources) {
				m.Character.Resources[m.selectedIndex].Use()
				return m, func() tea.Msg {
					return CharacterUpdatedMsg{Character: m.Character}
				}
			}
		case key.Matches(msg, m.keyMap.Restore):
			if m.selectedIndex < len(m.Character.Resources) {
				m.Character.Resources[m.selectedIndex].Restore()
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

// View implements tea.Model
func (m *StatsModel) View() string {
	var b strings.Builder

	// Stats section
	b.WriteString(styles.SectionStyle.Render("STATS") + "\n\n")

	stats := []struct {
		label string
		value string
	}{
		{"POISE", fmt.Sprintf("%d (%s)", m.Character.Stats.Poise, m.Character.Stats.PoiseDie)},
		{"Poise Points", fmt.Sprintf("%d", m.Character.Stats.PoisePoints)},
		{"Damage", fmt.Sprintf("%d - %s", m.Character.Stats.DamageTaken, m.Character.Stats.DamageNote)},
		{"Special", m.Character.Stats.Special},
		{"Parry & Repost", m.Character.Stats.ParryAndRepost},
		{"Mentor", m.Character.Mentor},
	}

	for _, stat := range stats {
		label := styles.LabelStyle.Render(stat.label)
		value := styles.ValueStyle.Render(stat.value)
		b.WriteString(fmt.Sprintf("  %s %s\n", label, value))
	}

	// Resources section
	b.WriteString("\n")
	b.WriteString(styles.SectionStyle.Render("RESOURCES") + "\n\n")

	for i, res := range m.Character.Resources {
		cursor := "  "
		nameStyle := styles.NormalStyle
		if i == m.selectedIndex {
			cursor = "> "
			nameStyle = styles.SelectedStyle
		}

		dots := styles.RenderResourceDots(res.Remaining, res.Total)
		name := nameStyle.Render(res.Name)
		count := styles.DimmedStyle.Render(fmt.Sprintf("(%d/%d)", res.Remaining, res.Total))

		b.WriteString(fmt.Sprintf("%s%s  %s %s\n", cursor, name, dots, count))
	}

	// Help for this view
	b.WriteString("\n")
	help := []string{
		styles.KeyStyle.Render("[j/k]") + " select",
		styles.KeyStyle.Render("[+/-]") + " use/restore",
		styles.KeyStyle.Render("[r]") + " rest",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}
