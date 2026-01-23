package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"sword-and-board/internal/adapters/tui/styles"
	"sword-and-board/internal/domain"
)

// SpellsKeyMap defines key bindings for the spells view
type SpellsKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Use    key.Binding
	Restore key.Binding
	Rest   key.Binding
	Add    key.Binding
	Delete key.Binding
	Enter  key.Binding
	Escape key.Binding
}

var DefaultSpellsKeyMap = SpellsKeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Use:     key.NewBinding(key.WithKeys("-")),
	Restore: key.NewBinding(key.WithKeys("+", "=")),
	Rest:    key.NewBinding(key.WithKeys("r")),
	Add:     key.NewBinding(key.WithKeys("a")),
	Delete:  key.NewBinding(key.WithKeys("d")),
	Enter:   key.NewBinding(key.WithKeys("enter")),
	Escape:  key.NewBinding(key.WithKeys("esc")),
}

// SpellsMode represents the current mode of the spells view
type SpellsMode int

const (
	SpellsModeNormal SpellsMode = iota
	SpellsModeAdd
	SpellsModeConfirmDelete
)

// SpellsModel handles the spells view
type SpellsModel struct {
	Character     *domain.Character
	selectedIndex int
	mode          SpellsMode
	keyMap        SpellsKeyMap

	// Input fields for add
	nameInput  textinput.Model
	usesInput  textinput.Model
	activeInput int
}

// NewSpellsModel creates a new spells view model
func NewSpellsModel(Character *domain.Character) *SpellsModel {
	m := &SpellsModel{
		Character: Character,
		keyMap:    DefaultSpellsKeyMap,
		mode:      SpellsModeNormal,
	}

	m.nameInput = textinput.New()
	m.nameInput.Placeholder = "Spell name"
	m.nameInput.CharLimit = 50

	m.usesInput = textinput.New()
	m.usesInput.Placeholder = "Total uses"
	m.usesInput.CharLimit = 3

	return m
}

// Init implements tea.Model
func (m *SpellsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *SpellsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case SpellsModeNormal:
			return m.updateNormal(msg)
		case SpellsModeAdd:
			return m.updateAddMode(msg)
		case SpellsModeConfirmDelete:
			return m.updateConfirmDelete(msg)
		}
	}

	return m, nil
}

func (m *SpellsModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Up):
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case key.Matches(msg, m.keyMap.Down):
		if m.selectedIndex < len(m.Character.Spells)-1 {
			m.selectedIndex++
		}
	case key.Matches(msg, m.keyMap.Use):
		if m.selectedIndex < len(m.Character.Spells) {
			if m.Character.Spells[m.selectedIndex].Use() {
				return m, func() tea.Msg {
					return CharacterUpdatedMsg{Character: m.Character}
				}
			}
			return m, func() tea.Msg {
				return StatusMsg{Message: "No uses remaining!", IsError: true}
			}
		}
	case key.Matches(msg, m.keyMap.Restore):
		if m.selectedIndex < len(m.Character.Spells) {
			if m.Character.Spells[m.selectedIndex].Restore() {
				return m, func() tea.Msg {
					return CharacterUpdatedMsg{Character: m.Character}
				}
			}
		}
	case key.Matches(msg, m.keyMap.Rest):
		m.Character.Rest()
		return m, func() tea.Msg {
			return StatusMsg{Message: "Rested at bonfire. All spells restored.", IsError: false}
		}
	case key.Matches(msg, m.keyMap.Add):
		m.mode = SpellsModeAdd
		m.clearInputs()
		m.nameInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keyMap.Delete):
		if m.selectedIndex < len(m.Character.Spells) {
			m.mode = SpellsModeConfirmDelete
		}
	}
	return m, nil
}

func (m *SpellsModel) updateAddMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Escape):
		m.mode = SpellsModeNormal
		return m, nil
	case key.Matches(msg, m.keyMap.Enter):
		if m.activeInput < 1 {
			m.activeInput++
			m.focusActiveInput()
			return m, textinput.Blink
		}
		return m.submitSpell()
	case msg.String() == "tab":
		m.activeInput = (m.activeInput + 1) % 2
		m.focusActiveInput()
		return m, textinput.Blink
	case msg.String() == "shift+tab":
		m.activeInput = (m.activeInput + 1) % 2
		m.focusActiveInput()
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	switch m.activeInput {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.usesInput, cmd = m.usesInput.Update(msg)
	}
	return m, cmd
}

func (m *SpellsModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.Character.RemoveSpell(m.selectedIndex)
		if m.selectedIndex >= len(m.Character.Spells) && m.selectedIndex > 0 {
			m.selectedIndex--
		}
		m.mode = SpellsModeNormal
		return m, func() tea.Msg {
			return CharacterUpdatedMsg{Character: m.Character}
		}
	case "n", "N", "esc":
		m.mode = SpellsModeNormal
	}
	return m, nil
}

func (m *SpellsModel) clearInputs() {
	m.nameInput.SetValue("")
	m.usesInput.SetValue("")
	m.activeInput = 0
}

func (m *SpellsModel) focusActiveInput() {
	m.nameInput.Blur()
	m.usesInput.Blur()

	switch m.activeInput {
	case 0:
		m.nameInput.Focus()
	case 1:
		m.usesInput.Focus()
	}
}

func (m *SpellsModel) submitSpell() (tea.Model, tea.Cmd) {
	name := m.nameInput.Value()
	if name == "" {
		return m, func() tea.Msg {
			return StatusMsg{Message: "Spell name is required", IsError: true}
		}
	}

	uses := 1
	if u := m.usesInput.Value(); u != "" {
		fmt.Sscanf(u, "%d", &uses)
	}

	spell := domain.Spell{
		Name:      name,
		TotalUses: uses,
		Used:      0,
	}

	m.Character.AddSpell(spell)
	m.selectedIndex = len(m.Character.Spells) - 1
	m.mode = SpellsModeNormal
	m.clearInputs()

	return m, func() tea.Msg {
		return CharacterUpdatedMsg{Character: m.Character}
	}
}

// View implements tea.Model
func (m *SpellsModel) View() string {
	var b strings.Builder

	if m.mode == SpellsModeAdd {
		return m.viewAddForm()
	}

	if m.mode == SpellsModeConfirmDelete {
		b.WriteString(styles.SectionStyle.Render("CONFIRM DELETE") + "\n\n")
		spell := m.Character.Spells[m.selectedIndex]
		b.WriteString(fmt.Sprintf("  Delete %s?\n\n", styles.SelectedStyle.Render(spell.Name)))
		b.WriteString("  " + styles.KeyStyle.Render("[y]") + " yes  ")
		b.WriteString(styles.KeyStyle.Render("[n]") + " no\n")
		return b.String()
	}

	b.WriteString(styles.SectionStyle.Render("SPELLS") + "\n\n")

	for i, spell := range m.Character.Spells {
		cursor := "  "
		nameStyle := styles.NormalStyle
		if i == m.selectedIndex {
			cursor = "> "
			nameStyle = styles.SelectedStyle
		}

		name := nameStyle.Render(fmt.Sprintf("%-18s", spell.Name))
		dots := styles.RenderSpellDots(spell.Used, spell.TotalUses)
		count := styles.DimmedStyle.Render(fmt.Sprintf("(%d/%d)", spell.Remaining(), spell.TotalUses))

		b.WriteString(fmt.Sprintf("%s%s %s %s\n", cursor, name, dots, count))
	}

	// Help
	b.WriteString("\n")
	help := []string{
		styles.KeyStyle.Render("[j/k]") + " select",
		styles.KeyStyle.Render("[+/-]") + " restore/use",
		styles.KeyStyle.Render("[r]") + " rest",
		styles.KeyStyle.Render("[a]") + " add",
		styles.KeyStyle.Render("[d]") + " delete",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}

func (m *SpellsModel) viewAddForm() string {
	var b strings.Builder

	b.WriteString(styles.SectionStyle.Render("ADD SPELL") + "\n\n")

	inputs := []struct {
		label string
		input textinput.Model
	}{
		{"Name", m.nameInput},
		{"Total Uses", m.usesInput},
	}

	for i, inp := range inputs {
		style := styles.NormalStyle
		if i == m.activeInput {
			style = styles.SelectedStyle
		}
		b.WriteString(fmt.Sprintf("  %s\n", style.Render(inp.label)))
		b.WriteString(fmt.Sprintf("  %s\n\n", inp.input.View()))
	}

	// Help
	help := []string{
		styles.KeyStyle.Render("[tab]") + " next field",
		styles.KeyStyle.Render("[enter]") + " submit",
		styles.KeyStyle.Render("[esc]") + " cancel",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}
