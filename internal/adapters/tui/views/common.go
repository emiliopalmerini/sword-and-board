package views

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"sword-and-board/internal/domain"
)

// ViewType represents the different views in the application
type ViewType int

const (
	ViewStats ViewType = iota
	ViewInventory
	ViewSpells
)

// String returns the display name of the view
func (v ViewType) String() string {
	switch v {
	case ViewStats:
		return "Stats"
	case ViewInventory:
		return "Inventory"
	case ViewSpells:
		return "Spells"
	default:
		return "Unknown"
	}
}

// Message types for inter-view communication

// CharacterUpdatedMsg is sent when the character data has been modified
type CharacterUpdatedMsg struct {
	Character *domain.Character
}

// SwitchViewMsg requests a view change
type SwitchViewMsg struct {
	View ViewType
}

// SaveMsg requests saving the character
type SaveMsg struct{}

// SavedMsg indicates the character was saved successfully
type SavedMsg struct{}

// SaveErrorMsg indicates a save error occurred
type SaveErrorMsg struct {
	Err error
}

// StatusMsg displays a temporary status message
type StatusMsg struct {
	Message string
	IsError bool
}

// ClearStatusMsg signals the App to clear the status line when its Gen matches the current generation.
type ClearStatusMsg struct {
	Gen int
}

func characterUpdatedCmd(character *domain.Character) tea.Cmd {
	return func() tea.Msg {
		return CharacterUpdatedMsg{Character: character}
	}
}

func statusCmd(message string, isError bool) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Message: message, IsError: isError}
	}
}

func parsePositiveInt(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}

	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("must be a positive number")
	}

	return n, nil
}

func parseNonNegativeInt(value string, fallback int) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}

	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("must be zero or a positive number")
	}

	return n, nil
}
