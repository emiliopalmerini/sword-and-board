package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sword-and-board/internal/adapters/tui/styles"
	"sword-and-board/internal/adapters/tui/views"
	"sword-and-board/internal/domain"
	"sword-and-board/internal/ports"
)

// statusDisplayDuration is how long a status message stays visible before auto-clearing.
const statusDisplayDuration = 3 * time.Second

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Stats     key.Binding
	Inventory key.Binding
	Spells    key.Binding
	NextView  key.Binding
	PrevView  key.Binding
	Save      key.Binding
	Quit      key.Binding
	Help      key.Binding
}

var DefaultKeyMap = KeyMap{
	Stats:     key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "stats")),
	Inventory: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "inventory")),
	Spells:    key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "spells")),
	NextView:  key.NewBinding(key.WithKeys("tab", "l"), key.WithHelp("tab/l", "next view")),
	PrevView:  key.NewBinding(key.WithKeys("shift+tab", "h"), key.WithHelp("shift+tab/h", "prev view")),
	Save:      key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "save")),
	Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
}

const numViews = 3

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
	statusGen   int
	showHelp    bool
	dirty       bool
	confirmQuit bool

	keyMap KeyMap
}

// sizedBox returns the box style sized to the current terminal, minus chrome
// reserved for header (4), tab bar (2), status (2), help (2), and box padding.
func (a *App) sizedBox() lipgloss.Style {
	box := styles.BoxStyle
	if a.width > 8 {
		box = box.Width(a.width - 4)
	}
	if a.height > 14 {
		box = box.Height(a.height - 12)
	}
	return box
}

// setStatus stores a message and returns a tea.Cmd that auto-clears it after statusDisplayDuration.
func (a *App) setStatus(msg string, isError bool) tea.Cmd {
	a.status = msg
	a.statusError = isError
	a.statusGen++
	gen := a.statusGen
	return tea.Tick(statusDisplayDuration, func(time.Time) tea.Msg {
		return views.ClearStatusMsg{Gen: gen}
	})
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
		// Quit confirmation prompt absorbs all keys until resolved.
		if a.confirmQuit {
			switch msg.String() {
			case "s":
				// Save then quit on success
				a.confirmQuit = false
				return a, tea.Sequence(a.save(), tea.Quit)
			case "d", "y":
				return a, tea.Quit
			case "n", "esc", "q":
				a.confirmQuit = false
				return a, nil
			}
			return a, nil
		}

		// Skip global key bindings when a sub-view is capturing text input
		if !a.isInputMode() {
			// Help modal dismissal — esc closes it, any view key still works.
			if a.showHelp && msg.String() == "esc" {
				a.showHelp = false
				return a, nil
			}

			if !a.showHelp && a.currentView == views.ViewStats && a.statsView.HandlesBeforeGlobal(msg) {
				newModel, cmd := a.statsView.Update(msg)
				a.statsView = newModel.(*views.StatsModel)
				return a, cmd
			}

			// Global key bindings
			switch {
			case key.Matches(msg, a.keyMap.Quit):
				if a.dirty {
					a.confirmQuit = true
					return a, nil
				}
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

			case key.Matches(msg, a.keyMap.NextView):
				a.currentView = views.ViewType((int(a.currentView) + 1) % numViews)
				return a, nil

			case key.Matches(msg, a.keyMap.PrevView):
				a.currentView = views.ViewType((int(a.currentView) + numViews - 1) % numViews)
				return a, nil

			case key.Matches(msg, a.keyMap.Save):
				return a, a.save()

			case key.Matches(msg, a.keyMap.Help):
				a.showHelp = !a.showHelp
				return a, nil
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case views.SavedMsg:
		a.dirty = false
		cmds = append(cmds, a.setStatus("Character saved!", false))

	case views.SaveErrorMsg:
		cmds = append(cmds, a.setStatus(fmt.Sprintf("Save failed: %v", msg.Err), true))

	case views.StatusMsg:
		cmds = append(cmds, a.setStatus(msg.Message, msg.IsError))

	case views.ClearStatusMsg:
		if msg.Gen == a.statusGen {
			a.status = ""
			a.statusError = false
		}

	case views.CharacterUpdatedMsg:
		// Character was updated, refresh views and mark dirty
		a.character = msg.Character
		a.statsView.Character = a.character
		a.inventoryView.Character = a.character
		a.spellsView.Character = a.character
		a.dirty = true
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

	// Header with character name and class; trailing * marks unsaved changes
	dirtyMark := ""
	if a.dirty {
		dirtyMark = styles.ErrorStyle.Render(" *")
	}
	header := styles.TitleStyle.Render("⚔  "+a.character.Name+"  ⚔") + dirtyMark
	subtitle := styles.SubtitleStyle.Render(
		fmt.Sprintf("%s • Path of the %s • Mentor: %s", a.character.Class, a.character.FaithPath, a.character.Mentor),
	)
	b.WriteString(header + "\n")
	b.WriteString(subtitle + "\n\n")

	// Tab bar with separator
	b.WriteString(a.renderTabs() + "\n")

	// Main content area — modal screens (help, quit confirm) replace the current view.
	var content string
	switch {
	case a.confirmQuit:
		content = a.renderConfirmQuit()
	case a.showHelp:
		content = a.renderFullHelp()
	default:
		switch a.currentView {
		case views.ViewStats:
			content = a.statsView.View()
		case views.ViewInventory:
			content = a.inventoryView.View()
		case views.ViewSpells:
			content = a.spellsView.View()
		}
	}

	b.WriteString(a.sizedBox().Render(a.centerIfModal(content)))

	// Status line
	if a.status != "" {
		b.WriteString("\n")
		if a.statusError {
			b.WriteString(styles.ErrorStyle.Render(a.status))
		} else {
			b.WriteString(styles.SuccessStyle.Render(a.status))
		}
	}

	// Help bar — short hints only; full help is now modal inside the box.
	b.WriteString("\n\n")
	b.WriteString(a.renderHelp())

	return b.String()
}

// renderConfirmQuit produces the unsaved-changes prompt shown as a modal.
func (a *App) renderConfirmQuit() string {
	var b strings.Builder
	b.WriteString(styles.SectionStyle.Render("UNSAVED CHANGES") + "\n\n")
	b.WriteString("  You have unsaved changes.\n\n")
	b.WriteString("  " + styles.KeyStyle.Render("[s]") + " save & quit\n")
	b.WriteString("  " + styles.KeyStyle.Render("[d]") + " discard & quit\n")
	b.WriteString("  " + styles.KeyStyle.Render("[n]") + " cancel\n")
	return b.String()
}

// centerIfModal vertically centers content within the box when a modal is active.
func (a *App) centerIfModal(content string) string {
	if !a.confirmQuit && !a.showHelp {
		return content
	}
	if a.height <= 14 {
		return content
	}
	innerHeight := a.height - 14
	return lipgloss.Place(a.width-8, innerHeight, lipgloss.Center, lipgloss.Center, content)
}

func (a *App) renderTabs() string {
	tabs := []string{"[1] Stats", "[2] Inventory", "[3] Spells"}
	sep := styles.TabSeparatorStyle.Render(" │ ")
	var rendered []string

	for i, tab := range tabs {
		if views.ViewType(i) == a.currentView {
			rendered = append(rendered, styles.ActiveTabStyle.Render(tab))
		} else {
			rendered = append(rendered, styles.InactiveTabStyle.Render(tab))
		}
	}

	return strings.Join(rendered, sep)
}

func (a *App) renderHelp() string {
	keys := []string{
		styles.KeyStyle.Render("[1/2/3]") + " view",
		styles.KeyStyle.Render("[tab]") + " next",
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
		{"tab / shift+tab", "Cycle views (next/prev)"},
		{"j/k or ↑/↓", "Navigate lists"},
		{"d / h", "Stats: take damage / heal"},
		{"enter / e", "Edit selected entry"},
		{"a", "Add resource/item/spell"},
		{"d", "Inventory/Spells: delete selected item"},
		{"x", "Stats: delete selected resource"},
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

// isInputMode returns true when the active sub-view is capturing text input
func (a *App) isInputMode() bool {
	switch a.currentView {
	case views.ViewStats:
		return a.statsView.IsInputMode()
	case views.ViewInventory:
		return a.inventoryView.IsInputMode()
	case views.ViewSpells:
		return a.spellsView.IsInputMode()
	default:
		return false
	}
}

func (a *App) save() tea.Cmd {
	return func() tea.Msg {
		if err := a.repository.Save(a.character); err != nil {
			return views.SaveErrorMsg{Err: err}
		}
		return views.SavedMsg{}
	}
}
