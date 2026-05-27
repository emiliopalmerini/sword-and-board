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
	Up        key.Binding
	Down      key.Binding
	Add       key.Binding
	Delete    key.Binding
	Edit      key.Binding
	Increment key.Binding
	Decrement key.Binding
	Enter     key.Binding
	Escape    key.Binding
}

var DefaultInventoryKeyMap = InventoryKeyMap{
	Up:        key.NewBinding(key.WithKeys("up", "k")),
	Down:      key.NewBinding(key.WithKeys("down", "j")),
	Add:       key.NewBinding(key.WithKeys("a")),
	Delete:    key.NewBinding(key.WithKeys("d")),
	Edit:      key.NewBinding(key.WithKeys("e")),
	Increment: key.NewBinding(key.WithKeys("+", "=")),
	Decrement: key.NewBinding(key.WithKeys("-")),
	Enter:     key.NewBinding(key.WithKeys("enter")),
	Escape:    key.NewBinding(key.WithKeys("esc")),
}

// InventoryMode represents the current mode of the inventory view
type InventoryMode int

const (
	InventoryModeNormal InventoryMode = iota
	InventoryModeAdd
	InventoryModeEdit
	InventoryModeConfirmDelete
)

// itemTypeCycle is the cycle order for the type field in the add/edit form.
var itemTypeCycle = []domain.ItemType{
	domain.ItemTypeWeapon,
	domain.ItemTypeEquipment,
	domain.ItemTypeConsumable,
}

// indexOfItemType returns the cycle index for an item type, or 1 (Equipment) if unknown.
func indexOfItemType(t domain.ItemType) int {
	for i, it := range itemTypeCycle {
		if it == t {
			return i
		}
	}
	return 1
}

// InventoryModel handles the inventory view
type InventoryModel struct {
	Character     *domain.Character
	selectedIndex int
	mode          InventoryMode
	keyMap        InventoryKeyMap

	// Input fields for add/edit
	nameInput     textinput.Model
	quantityInput textinput.Model
	typeIndex     int
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
			m.typeIndex = indexOfItemType(item.Type)
			m.notesInput.SetValue(item.Notes)
			m.nameInput.Focus()
			return m, textinput.Blink
		}
	case key.Matches(msg, m.keyMap.Increment):
		return m.adjustSelectedQuantity(1)
	case key.Matches(msg, m.keyMap.Decrement):
		return m.adjustSelectedQuantity(-1)
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
			m.activeInput++
			m.focusActiveInput()
			return m, textinput.Blink
		}
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

	// Type field cycles with left/right; other fields use textinput.
	if m.activeInput == 2 {
		switch msg.String() {
		case "left", "h":
			m.typeIndex = (m.typeIndex + len(itemTypeCycle) - 1) % len(itemTypeCycle)
		case "right", "l", " ":
			m.typeIndex = (m.typeIndex + 1) % len(itemTypeCycle)
		}
		return m, nil
	}

	var cmd tea.Cmd
	switch m.activeInput {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.quantityInput, cmd = m.quantityInput.Update(msg)
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
		return m, characterUpdatedCmd(m.Character)
	case "n", "N", "esc":
		m.mode = InventoryModeNormal
	}
	return m, nil
}

func (m *InventoryModel) adjustSelectedQuantity(delta int) (tea.Model, tea.Cmd) {
	if m.selectedIndex < 0 || m.selectedIndex >= len(m.Character.Inventory) {
		return m, nil
	}

	item := &m.Character.Inventory[m.selectedIndex]
	if delta < 0 && item.Quantity <= 1 {
		return m, statusCmd("Quantity is already 1. Press [d] to delete.", true)
	}

	item.Quantity += delta
	return m, tea.Batch(
		characterUpdatedCmd(m.Character),
		statusCmd(fmt.Sprintf("%s quantity: %d", item.Name, item.Quantity), false),
	)
}

func (m *InventoryModel) clearInputs() {
	m.nameInput.SetValue("")
	m.quantityInput.SetValue("")
	m.typeIndex = 1 // default Equipment
	m.notesInput.SetValue("")
	m.activeInput = 0
}

func (m *InventoryModel) focusActiveInput() {
	m.nameInput.Blur()
	m.quantityInput.Blur()
	m.notesInput.Blur()

	switch m.activeInput {
	case 0:
		m.nameInput.Focus()
	case 1:
		m.quantityInput.Focus()
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

	quantity, err := parsePositiveInt(m.quantityInput.Value(), 1)
	if err != nil {
		return m, statusCmd("Quantity must be a positive number", true)
	}

	itemType := domain.ItemTypeEquipment
	if m.typeIndex >= 0 && m.typeIndex < len(itemTypeCycle) {
		itemType = itemTypeCycle[m.typeIndex]
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

	return m, characterUpdatedCmd(m.Character)
}

// IsInputMode returns true when the view is capturing text input
func (m *InventoryModel) IsInputMode() bool {
	return m.mode == InventoryModeAdd || m.mode == InventoryModeEdit
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

	if len(m.Character.Inventory) == 0 {
		b.WriteString(styles.EmptyStateStyle.Render("No items. Press [a] to add one.") + "\n\n")

		help := []string{
			styles.KeyStyle.Render("[a]") + " add",
		}
		b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))
		return b.String()
	}

	// Group items by type while preserving original indexes for edit/delete.
	itemsByType := m.inventoryRowsByType()
	typeOrder := []domain.ItemType{domain.ItemTypeWeapon, domain.ItemTypeEquipment, domain.ItemTypeConsumable}

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
		for _, row := range items {
			item := row.item
			itemIdx := row.index

			cursor := "    "
			nameStyle := styles.NormalStyle
			if itemIdx == m.selectedIndex {
				cursor = "  > "
				nameStyle = styles.SelectedStyle
			}

			name := nameStyle.Render(item.Name)
			qty := ""
			if item.Quantity > 1 {
				qty = " " + styles.DimmedStyle.Render(fmt.Sprintf("x%d", item.Quantity))
			}
			notes := ""
			if item.Notes != "" {
				notes = " " + styles.DimmedStyle.Render("— "+item.Notes)
			}
			b.WriteString(fmt.Sprintf("%s%s%s%s\n", cursor, name, qty, notes))
		}
		b.WriteString("\n")
	}

	// Help
	help := []string{
		styles.KeyStyle.Render("[j/k]") + " select",
		styles.KeyStyle.Render("[a]") + " add",
		styles.KeyStyle.Render("[e]") + " edit",
		styles.KeyStyle.Render("[+/-]") + " qty",
		styles.KeyStyle.Render("[d]") + " delete",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}

type inventoryRow struct {
	index int
	item  domain.Item
}

func (m *InventoryModel) inventoryRowsByType() map[domain.ItemType][]inventoryRow {
	result := make(map[domain.ItemType][]inventoryRow)
	for i, item := range m.Character.Inventory {
		result[item.Type] = append(result[item.Type], inventoryRow{
			index: i,
			item:  item,
		})
	}
	return result
}

func (m *InventoryModel) viewInputForm() string {
	var b strings.Builder

	title := "ADD ITEM"
	if m.mode == InventoryModeEdit {
		title = "EDIT ITEM"
	}
	b.WriteString(styles.SectionStyle.Render(title) + "\n\n")

	labels := []string{"Name", "Quantity", "Type", "Notes"}
	for i, label := range labels {
		style := styles.NormalStyle
		if i == m.activeInput {
			style = styles.SelectedStyle
		}
		b.WriteString(fmt.Sprintf("  %s\n", style.Render(label)))

		var renderedField string
		switch i {
		case 0:
			renderedField = m.nameInput.View()
		case 1:
			renderedField = m.quantityInput.View()
		case 2:
			renderedField = m.renderTypeField()
		case 3:
			renderedField = m.notesInput.View()
		}
		b.WriteString(fmt.Sprintf("  %s\n\n", renderedField))
	}

	help := []string{
		styles.KeyStyle.Render("[tab]") + " next field",
		styles.KeyStyle.Render("[←/→]") + " cycle type",
		styles.KeyStyle.Render("[enter]") + " submit",
		styles.KeyStyle.Render("[esc]") + " cancel",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}

// renderTypeField shows the current type with neighbors visible to suggest cycling.
func (m *InventoryModel) renderTypeField() string {
	var parts []string
	for i, t := range itemTypeCycle {
		var styled string
		if i == m.typeIndex {
			var colored string
			switch t {
			case domain.ItemTypeWeapon:
				colored = styles.WeaponStyle.Render(string(t))
			case domain.ItemTypeEquipment:
				colored = styles.EquipmentStyle.Render(string(t))
			case domain.ItemTypeConsumable:
				colored = styles.ConsumableStyle.Render(string(t))
			default:
				colored = string(t)
			}
			styled = styles.SelectedStyle.Render("‹ ") + colored + styles.SelectedStyle.Render(" ›")
		} else {
			styled = styles.DimmedStyle.Render(string(t))
		}
		parts = append(parts, styled)
	}
	return strings.Join(parts, "  ")
}
