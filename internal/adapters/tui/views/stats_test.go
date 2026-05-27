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

func TestEditProfileField(t *testing.T) {
	character := &domain.Character{Name: "Sir Dissimotis"}
	model := NewStatsModel(character)

	model = updateStatsModel(t, model, keyRune('e'))
	if !model.editing {
		t.Fatal("expected field editor to be open")
	}

	model.editInput.SetValue("Lady Dissimotis")
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))

	if got := character.Name; got != "Lady Dissimotis" {
		t.Fatalf("name = %q, want %q", got, "Lady Dissimotis")
	}
}

func TestAddResource(t *testing.T) {
	character := &domain.Character{}
	model := NewStatsModel(character)

	model = updateStatsModel(t, model, keyRune('a'))
	if !model.editingResource {
		t.Fatal("expected resource form to be open")
	}

	model.resourceNameInput.SetValue("Estus Flask")
	model.resourceTotalInput.SetValue("2")
	model.resourceRemainingInput.SetValue("1")
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))

	if got := len(character.Resources); got != 1 {
		t.Fatalf("resource count = %d, want 1", got)
	}
	resource := character.Resources[0]
	if resource.Name != "Estus Flask" || resource.Total != 2 || resource.Remaining != 1 {
		t.Fatalf("resource = %+v, want Estus Flask 1/2", resource)
	}
}

func TestEditSelectedResource(t *testing.T) {
	character := &domain.Character{
		Resources: []domain.Resource{
			{Name: "Estus Flask", Total: 2, Remaining: 1},
		},
	}
	model := NewStatsModel(character)
	model.selectedIndex = len(model.statsFields())

	model = updateStatsModel(t, model, keyRune('e'))
	if !model.editingResource {
		t.Fatal("expected resource form to be open")
	}

	model.resourceNameInput.SetValue("Ashen Estus")
	model.resourceTotalInput.SetValue("3")
	model.resourceRemainingInput.SetValue("2")
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))
	model = updateStatsModel(t, model, keyType(tea.KeyEnter))

	resource := character.Resources[0]
	if resource.Name != "Ashen Estus" || resource.Total != 3 || resource.Remaining != 2 {
		t.Fatalf("resource = %+v, want Ashen Estus 2/3", resource)
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
