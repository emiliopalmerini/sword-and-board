package yaml

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"sword-and-board/internal/domain"
)

// Repository implements CharacterRepository using YAML file storage
type Repository struct {
	filePath string
}

// NewRepository creates a new YAML repository
func NewRepository(filePath string) *Repository {
	return &Repository{filePath: filePath}
}

// Load reads the character from the YAML file
func (r *Repository) Load() (*domain.Character, error) {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var character domain.Character
	if err := yaml.Unmarshal(data, &character); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &character, nil
}

// Save writes the character to the YAML file
func (r *Repository) Save(character *domain.Character) error {
	data, err := yaml.Marshal(character)
	if err != nil {
		return fmt.Errorf("failed to marshal character: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
