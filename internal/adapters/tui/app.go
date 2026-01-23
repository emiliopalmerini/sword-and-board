package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sword-and-board/internal/adapters/tui/styles"
	"sword-and-board/internal/adapters/tui/views"
	"sword-and-board/internal/domain"
	"sword-and-board/internal/ports"
)

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Stats     key.Binding
	Inventory key.Binding
	Spells    key.Binding
	Save      key.Binding
	Quit      key.Binding
	Help      key.Binding
}

var DefaultKeyMap = KeyMap{
	Stats:     key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "stats")),
	Inventory: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "inventory")),
	Spells:    key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "spells")),
	Save:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "save")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}

// App is the main application model
type App struct {
	character   *domain.Character
	repository  ports.CharacterRepository
	currentView views.ViewType

	// Sub-models for each view
	statsView     *views.StatsModel
	inventoryView *views.InventoryModel
	spellsView    *views.SpellsModel

	// UI state
	width       int
	height      int
	status      string
	statusError bool
	showHelp    bool

	keyMap KeyMap
}

// NewApp creates a new application instance
func NewApp(character *domain.Character, repo ports.CharacterRepository) *App {
	app := &App{
		character:   character,
		repository:  repo,
		currentView: views.ViewStats,
		keyMap:      DefaultKeyMap,
	}

	// Initialize sub-models
	app.statsView = views.NewStatsModel(character)
	app.inventoryView = views.NewInventoryModel(character)
	app.spellsView = views.NewSpellsModel(character)

	return app
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear status on any key press
		a.status = ""
		a.statusError = false

		// Global key bindings
		switch {
		case key.Matches(msg, a.keyMap.Quit):
			return a, tea.Quit

		case key.Matches(msg, a.keyMap.Stats):
			a.currentView = views.ViewStats
			return a, nil

		case key.Matches(msg, a.keyMap.Inventory):
			a.currentView = views.ViewInventory
			return a, nil

		case key.Matches(msg, a.keyMap.Spells):
			a.currentView = views.ViewSpells
			return a, nil

		case key.Matches(msg, a.keyMap.Save):
			return a, a.save()

		case key.Matches(msg, a.keyMap.Help):
			a.showHelp = !a.showHelp
			return a, nil
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case views.SavedMsg:
		a.status = "Character saved!"
		a.statusError = false

	case views.SaveErrorMsg:
		a.status = fmt.Sprintf("Save failed: %v", msg.Err)
		a.statusError = true

	case views.StatusMsg:
		a.status = msg.Message
		a.statusError = msg.IsError

	case views.CharacterUpdatedMsg:
		// Character was updated, refresh views
		a.character = msg.Character
		a.statsView.Character = a.character
		a.inventoryView.Character = a.character
		a.spellsView.Character = a.character
	}

	// Route to current view
	switch a.currentView {
	case views.ViewStats:
		newModel, cmd := a.statsView.Update(msg)
		a.statsView = newModel.(*views.StatsModel)
		cmds = append(cmds, cmd)

	case views.ViewInventory:
		newModel, cmd := a.inventoryView.Update(msg)
		a.inventoryView = newModel.(*views.InventoryModel)
		cmds = append(cmds, cmd)

	case views.ViewSpells:
		newModel, cmd := a.spellsView.Update(msg)
		a.spellsView = newModel.(*views.SpellsModel)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

// View implements tea.Model
func (a *App) View() string {
	var b strings.Builder

	// Header with character name
	header := styles.TitleStyle.Render(a.character.Name)
	subtitle := styles.SubtitleStyle.Render(
		fmt.Sprintf("%s • Path of the %s", a.character.Class, a.character.FaithPath),
	)
	b.WriteString(header + "\n")
	b.WriteString(subtitle + "\n\n")

	// Tab bar
	b.WriteString(a.renderTabs() + "\n\n")

	// Current view content
	switch a.currentView {
	case views.ViewStats:
		b.WriteString(a.statsView.View())
	case views.ViewInventory:
		b.WriteString(a.inventoryView.View())
	case views.ViewSpells:
		b.WriteString(a.spellsView.View())
	}

	// Status line
	if a.status != "" {
		b.WriteString("\n")
		if a.statusError {
			b.WriteString(styles.ErrorStyle.Render(a.status))
		} else {
			b.WriteString(styles.SuccessStyle.Render(a.status))
		}
	}

	// Help bar
	b.WriteString("\n\n")
	if a.showHelp {
		b.WriteString(a.renderFullHelp())
	} else {
		b.WriteString(a.renderHelp())
	}

	return b.String()
}

func (a *App) renderTabs() string {
	tabs := []string{"[1] Stats", "[2] Inventory", "[3] Spells"}
	var rendered []string

	for i, tab := range tabs {
		if views.ViewType(i) == a.currentView {
			rendered = append(rendered, styles.ActiveTabStyle.Render(tab))
		} else {
			rendered = append(rendered, styles.InactiveTabStyle.Render(tab))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func (a *App) renderHelp() string {
	keys := []string{
		styles.KeyStyle.Render("[1/2/3]") + " switch view",
		styles.KeyStyle.Render("[s]") + " save",
		styles.KeyStyle.Render("[?]") + " help",
		styles.KeyStyle.Render("[q]") + " quit",
	}
	return styles.HelpStyle.Render(strings.Join(keys, "  "))
}

func (a *App) renderFullHelp() string {
	var b strings.Builder
	b.WriteString(styles.SectionStyle.Render("KEYBOARD SHORTCUTS") + "\n\n")

	help := []struct {
		key  string
		desc string
	}{
		{"1/2/3", "Switch between Stats/Inventory/Spells"},
		{"j/k or ↑/↓", "Navigate lists"},
		{"e", "Edit selected item"},
		{"a", "Add new item"},
		{"d", "Delete selected item"},
		{"+/-", "Use/restore spell or resource"},
		{"r", "Rest (restore all)"},
		{"s", "Save character"},
		{"?", "Toggle help"},
		{"q", "Quit"},
	}

	for _, h := range help {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			styles.KeyStyle.Render(fmt.Sprintf("%-12s", h.key)),
			h.desc,
		))
	}

	return b.String()
}

func (a *App) save() tea.Cmd {
	return func() tea.Msg {
		if err := a.repository.Save(a.character); err != nil {
			return views.SaveErrorMsg{Err: err}
		}
		return views.SavedMsg{}
	}
}
