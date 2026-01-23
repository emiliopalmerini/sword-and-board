package views

import (
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
