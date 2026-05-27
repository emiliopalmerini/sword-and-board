package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"sword-and-board/internal/domain"
)

func TestDamagePromptUsesTypedAmount(t *testing.T) {
	character := &domain.Character{
		Stats: domain.Stats{MaxHP: 10},
	}
	model := NewStatsModel(character)

	model = updateStatsModel(t, model, keyRune('d'))
	if !model.editingDamage {
		t.Fatal("expected damage prompt to be open")
	}

	model = updateStatsModel(t, model, keyRune('3'))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))

	if got := character.Stats.DamageTaken; got != 3 {
		t.Fatalf("damage taken = %d, want 3", got)
	}
}

func TestDamagePromptDefaultsToOne(t *testing.T) {
	character := &domain.Character{
		Stats: domain.Stats{MaxHP: 10},
	}
	model := NewStatsModel(character)

	model = updateStatsModel(t, model, keyRune('d'))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))

	if got := character.Stats.DamageTaken; got != 1 {
		t.Fatalf("damage taken = %d, want 1", got)
	}
}

func TestHealPromptDefaultsToOne(t *testing.T) {
	character := &domain.Character{
		Stats: domain.Stats{MaxHP: 10, DamageTaken: 3},
	}
	model := NewStatsModel(character)

	model = updateStatsModel(t, model, keyRune('h'))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))

	if got := character.Stats.DamageTaken; got != 2 {
		t.Fatalf("damage taken = %d, want 2", got)
	}
}

func TestHealPromptDoesNotChangeFullHP(t *testing.T) {
	character := &domain.Character{
		Stats: domain.Stats{MaxHP: 10},
	}
	model := NewStatsModel(character)

	model = updateStatsModel(t, model, keyRune('h'))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))

	if got := character.Stats.DamageTaken; got != 0 {
		t.Fatalf("damage taken = %d, want 0", got)
	}
	if model.editingDamage {
		t.Fatal("expected damage prompt to close")
	}
}

func updateStatsModel(t *testing.T, model *StatsModel, msg tea.Msg) *StatsModel {
	t.Helper()

	updated, _ := model.Update(msg)
	statsModel, ok := updated.(*StatsModel)
	if !ok {
		t.Fatalf("updated model type = %T, want *StatsModel", updated)
	}

	return statsModel
}

func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func keyType(keyType tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: keyType}
}
