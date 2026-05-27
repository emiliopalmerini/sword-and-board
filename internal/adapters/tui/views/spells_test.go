package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"sword-and-board/internal/domain"
)

func TestEditSpellUsedCount(t *testing.T) {
	character := &domain.Character{
		Spells: []domain.Spell{
			{Name: "Lightning", TotalUses: 10, Used: 4},
		},
	}
	model := NewSpellsModel(character)

	model = updateSpellsModel(t, model, keyRune('e'))
	if model.mode != SpellsModeEdit {
		t.Fatal("expected spell edit form to be open")
	}

	model.nameInput.SetValue("Lightning Spear")
	model.usesInput.SetValue("8")
	model.usedInput.SetValue("3")
	model = updateSpellsModel(t, model, keyType(tea.KeyEnter))
	model = updateSpellsModel(t, model, keyType(tea.KeyEnter))
	model = updateSpellsModel(t, model, keyType(tea.KeyEnter))

	spell := character.Spells[0]
	if spell.Name != "Lightning Spear" || spell.TotalUses != 8 || spell.Used != 3 {
		t.Fatalf("spell = %+v, want Lightning Spear 3/8 used", spell)
	}
}

func updateSpellsModel(t *testing.T, model *SpellsModel, msg tea.Msg) *SpellsModel {
	t.Helper()

	updated, _ := model.Update(msg)
	spellsModel, ok := updated.(*SpellsModel)
	if !ok {
		t.Fatalf("updated model type = %T, want *SpellsModel", updated)
	}

	return spellsModel
}
