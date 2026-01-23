package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"sword-and-board/internal/adapters/tui"
	yamlrepo "sword-and-board/internal/adapters/yaml"
)

func main() {
	// Find the data file
	dataPath := findDataFile()

	// Load character from YAML
	repo := yamlrepo.NewRepository(dataPath)
	character, err := repo.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading character: %v\n", err)
		os.Exit(1)
	}

	// Create and run the TUI
	app := tui.NewApp(character, repo)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

// findDataFile looks for the character data file in common locations
func findDataFile() string {
	// Check common locations in order of preference
	locations := []string{
		"data/dissimotis.yaml",                    // Current directory
		"./data/dissimotis.yaml",                  // Explicit current directory
		filepath.Join(getExecutableDir(), "data", "dissimotis.yaml"), // Next to executable
	}

	// Also check $HOME/.config/sword-and-board/
	if home, err := os.UserHomeDir(); err == nil {
		locations = append(locations, filepath.Join(home, ".config", "sword-and-board", "dissimotis.yaml"))
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	// Default to the first location (will error later if not found)
	return locations[0]
}

// getExecutableDir returns the directory containing the executable
func getExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}
	return filepath.Dir(exe)
}
