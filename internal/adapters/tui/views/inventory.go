package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"sword-and-board/internal/adapters/tui/styles"
	"sword-and-board/internal/domain"
)

// InventoryKeyMap defines key bindings for the inventory view
type InventoryKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Add    key.Binding
	Delete key.Binding
	Edit   key.Binding
	Enter  key.Binding
	Escape key.Binding
}

var DefaultInventoryKeyMap = InventoryKeyMap{
	Up:     key.NewBinding(key.WithKeys("up", "k")),
	Down:   key.NewBinding(key.WithKeys("down", "j")),
	Add:    key.NewBinding(key.WithKeys("a")),
	Delete: key.NewBinding(key.WithKeys("d")),
	Edit:   key.NewBinding(key.WithKeys("e")),
	Enter:  key.NewBinding(key.WithKeys("enter")),
	Escape: key.NewBinding(key.WithKeys("esc")),
}

// InventoryMode represents the current mode of the inventory view
type InventoryMode int

const (
	InventoryModeNormal InventoryMode = iota
	InventoryModeAdd
	InventoryModeEdit
	InventoryModeConfirmDelete
)

// InventoryModel handles the inventory view
type InventoryModel struct {
	Character     *domain.Character
	selectedIndex int
	mode          InventoryMode
	keyMap        InventoryKeyMap

	// Input fields for add/edit
	nameInput     textinput.Model
	quantityInput textinput.Model
	typeInput     textinput.Model
	notesInput    textinput.Model
	activeInput   int
}

// NewInventoryModel creates a new inventory view model
func NewInventoryModel(Character *domain.Character) *InventoryModel {
	m := &InventoryModel{
		Character: Character,
		keyMap:    DefaultInventoryKeyMap,
		mode:      InventoryModeNormal,
	}

	// Initialize text inputs
	m.nameInput = textinput.New()
	m.nameInput.Placeholder = "Item name"
	m.nameInput.CharLimit = 50

	m.quantityInput = textinput.New()
	m.quantityInput.Placeholder = "1"
	m.quantityInput.CharLimit = 3

	m.typeInput = textinput.New()
	m.typeInput.Placeholder = "Weapon/Equipment/Consumable"
	m.typeInput.CharLimit = 20

	m.notesInput = textinput.New()
	m.notesInput.Placeholder = "Notes (optional)"
	m.notesInput.CharLimit = 100

	return m
}

// Init implements tea.Model
func (m *InventoryModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *InventoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case InventoryModeNormal:
			return m.updateNormal(msg)
		case InventoryModeAdd, InventoryModeEdit:
			return m.updateInputMode(msg)
		case InventoryModeConfirmDelete:
			return m.updateConfirmDelete(msg)
		}
	}

	return m, nil
}

func (m *InventoryModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Up):
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}
	case key.Matches(msg, m.keyMap.Down):
		if m.selectedIndex < len(m.Character.Inventory)-1 {
			m.selectedIndex++
		}
	case key.Matches(msg, m.keyMap.Add):
		m.mode = InventoryModeAdd
		m.clearInputs()
		m.nameInput.Focus()
		return m, textinput.Blink
	case key.Matches(msg, m.keyMap.Edit):
		if m.selectedIndex < len(m.Character.Inventory) {
			m.mode = InventoryModeEdit
			item := m.Character.Inventory[m.selectedIndex]
			m.nameInput.SetValue(item.Name)
			m.quantityInput.SetValue(fmt.Sprintf("%d", item.Quantity))
			m.typeInput.SetValue(string(item.Type))
			m.notesInput.SetValue(item.Notes)
			m.nameInput.Focus()
			return m, textinput.Blink
		}
	case key.Matches(msg, m.keyMap.Delete):
		if m.selectedIndex < len(m.Character.Inventory) {
			m.mode = InventoryModeConfirmDelete
		}
	}
	return m, nil
}

func (m *InventoryModel) updateInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Escape):
		m.mode = InventoryModeNormal
		return m, nil
	case key.Matches(msg, m.keyMap.Enter):
		if m.activeInput < 3 {
			// Move to next input
			m.activeInput++
			m.focusActiveInput()
			return m, textinput.Blink
		}
		// Submit
		return m.submitItem()
	case msg.String() == "tab":
		m.activeInput = (m.activeInput + 1) % 4
		m.focusActiveInput()
		return m, textinput.Blink
	case msg.String() == "shift+tab":
		m.activeInput = (m.activeInput + 3) % 4
		m.focusActiveInput()
		return m, textinput.Blink
	}

	// Update the active input
	var cmd tea.Cmd
	switch m.activeInput {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.quantityInput, cmd = m.quantityInput.Update(msg)
	case 2:
		m.typeInput, cmd = m.typeInput.Update(msg)
	case 3:
		m.notesInput, cmd = m.notesInput.Update(msg)
	}
	return m, cmd
}

func (m *InventoryModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.Character.RemoveItem(m.selectedIndex)
		if m.selectedIndex >= len(m.Character.Inventory) && m.selectedIndex > 0 {
			m.selectedIndex--
		}
		m.mode = InventoryModeNormal
		return m, func() tea.Msg {
			return CharacterUpdatedMsg{Character: m.Character}
		}
	case "n", "N", "esc":
		m.mode = InventoryModeNormal
	}
	return m, nil
}

func (m *InventoryModel) clearInputs() {
	m.nameInput.SetValue("")
	m.quantityInput.SetValue("")
	m.typeInput.SetValue("")
	m.notesInput.SetValue("")
	m.activeInput = 0
}

func (m *InventoryModel) focusActiveInput() {
	m.nameInput.Blur()
	m.quantityInput.Blur()
	m.typeInput.Blur()
	m.notesInput.Blur()

	switch m.activeInput {
	case 0:
		m.nameInput.Focus()
	case 1:
		m.quantityInput.Focus()
	case 2:
		m.typeInput.Focus()
	case 3:
		m.notesInput.Focus()
	}
}

func (m *InventoryModel) submitItem() (tea.Model, tea.Cmd) {
	name := m.nameInput.Value()
	if name == "" {
		return m, func() tea.Msg {
			return StatusMsg{Message: "Item name is required", IsError: true}
		}
	}

	quantity := 1
	if q := m.quantityInput.Value(); q != "" {
		fmt.Sscanf(q, "%d", &quantity)
	}

	itemType := domain.ItemTypeEquipment
	switch strings.ToLower(m.typeInput.Value()) {
	case "weapon", "w":
		itemType = domain.ItemTypeWeapon
	case "consumable", "c":
		itemType = domain.ItemTypeConsumable
	}

	item := domain.Item{
		Name:     name,
		Quantity: quantity,
		Type:     itemType,
		Notes:    m.notesInput.Value(),
	}

	if m.mode == InventoryModeAdd {
		m.Character.AddItem(item)
		m.selectedIndex = len(m.Character.Inventory) - 1
	} else {
		m.Character.Inventory[m.selectedIndex] = item
	}

	m.mode = InventoryModeNormal
	m.clearInputs()

	return m, func() tea.Msg {
		return CharacterUpdatedMsg{Character: m.Character}
	}
}

// View implements tea.Model
func (m *InventoryModel) View() string {
	var b strings.Builder

	if m.mode == InventoryModeAdd || m.mode == InventoryModeEdit {
		return m.viewInputForm()
	}

	if m.mode == InventoryModeConfirmDelete {
		b.WriteString(styles.SectionStyle.Render("CONFIRM DELETE") + "\n\n")
		item := m.Character.Inventory[m.selectedIndex]
		b.WriteString(fmt.Sprintf("  Delete %s?\n\n", styles.SelectedStyle.Render(item.Name)))
		b.WriteString("  " + styles.KeyStyle.Render("[y]") + " yes  ")
		b.WriteString(styles.KeyStyle.Render("[n]") + " no\n")
		return b.String()
	}

	b.WriteString(styles.SectionStyle.Render("INVENTORY") + "\n\n")

	// Group items by type
	itemsByType := m.Character.ItemsByType()
	typeOrder := []domain.ItemType{domain.ItemTypeWeapon, domain.ItemTypeEquipment, domain.ItemTypeConsumable}

	currentIdx := 0
	for _, itemType := range typeOrder {
		items, ok := itemsByType[itemType]
		if !ok || len(items) == 0 {
			continue
		}

		// Type header
		typeStyle := styles.NormalStyle
		switch itemType {
		case domain.ItemTypeWeapon:
			typeStyle = styles.WeaponStyle
		case domain.ItemTypeEquipment:
			typeStyle = styles.EquipmentStyle
		case domain.ItemTypeConsumable:
			typeStyle = styles.ConsumableStyle
		}
		b.WriteString("  " + typeStyle.Render(string(itemType)+"s") + "\n")

		// Items in this type
		for _, item := range items {
			// Find this item's index in the main inventory
			itemIdx := -1
			for i, invItem := range m.Character.Inventory {
				if invItem.Name == item.Name && invItem.Type == item.Type {
					itemIdx = i
					break
				}
			}

			cursor := "    "
			nameStyle := styles.NormalStyle
			if itemIdx == m.selectedIndex {
				cursor = "  > "
				nameStyle = styles.SelectedStyle
			}

			name := nameStyle.Render(item.Name)
			qty := styles.DimmedStyle.Render(fmt.Sprintf("(%d)", item.Quantity))
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, name, qty))

			if item.Notes != "" && itemIdx == m.selectedIndex {
				b.WriteString("      " + styles.DimmedStyle.Render(item.Notes) + "\n")
			}

			currentIdx++
		}
		b.WriteString("\n")
	}

	// Help
	help := []string{
		styles.KeyStyle.Render("[j/k]") + " select",
		styles.KeyStyle.Render("[a]") + " add",
		styles.KeyStyle.Render("[e]") + " edit",
		styles.KeyStyle.Render("[d]") + " delete",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}

func (m *InventoryModel) viewInputForm() string {
	var b strings.Builder

	title := "ADD ITEM"
	if m.mode == InventoryModeEdit {
		title = "EDIT ITEM"
	}
	b.WriteString(styles.SectionStyle.Render(title) + "\n\n")

	inputs := []struct {
		label string
		input textinput.Model
	}{
		{"Name", m.nameInput},
		{"Quantity", m.quantityInput},
		{"Type", m.typeInput},
		{"Notes", m.notesInput},
	}

	for i, inp := range inputs {
		style := styles.NormalStyle
		if i == m.activeInput {
			style = styles.SelectedStyle
		}
		b.WriteString(fmt.Sprintf("  %s\n", style.Render(inp.label)))
		b.WriteString(fmt.Sprintf("  %s\n\n", inp.input.View()))
	}

	// Help
	help := []string{
		styles.KeyStyle.Render("[tab]") + " next field",
		styles.KeyStyle.Render("[enter]") + " submit",
		styles.KeyStyle.Render("[esc]") + " cancel",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}
