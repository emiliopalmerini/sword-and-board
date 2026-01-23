package ports

import "sword-and-board/internal/domain"

// CharacterRepository defines the interface for loading and saving character data
type CharacterRepository interface {
	// Load reads the character from storage
	Load() (*domain.Character, error)

	// Save writes the character to storage
	Save(character *domain.Character) error
}
