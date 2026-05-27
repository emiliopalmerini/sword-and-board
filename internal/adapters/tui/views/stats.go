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

// StatsKeyMap defines key bindings for the stats view
type StatsKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Damage  key.Binding
	Heal    key.Binding
	Use     key.Binding
	Restore key.Binding
	Rest    key.Binding
	Add     key.Binding
	Delete  key.Binding
	Edit    key.Binding
	Enter   key.Binding
	Escape  key.Binding
}

var DefaultStatsKeyMap = StatsKeyMap{
	Up:      key.NewBinding(key.WithKeys("up", "k")),
	Down:    key.NewBinding(key.WithKeys("down", "j")),
	Damage:  key.NewBinding(key.WithKeys("d")),
	Heal:    key.NewBinding(key.WithKeys("h", "H")),
	Use:     key.NewBinding(key.WithKeys("-")),
	Restore: key.NewBinding(key.WithKeys("+", "=")),
	Rest:    key.NewBinding(key.WithKeys("r")),
	Add:     key.NewBinding(key.WithKeys("a")),
	Delete:  key.NewBinding(key.WithKeys("x")),
	Edit:    key.NewBinding(key.WithKeys("e")),
	Enter:   key.NewBinding(key.WithKeys("enter")),
	Escape:  key.NewBinding(key.WithKeys("esc")),
}

// StatsModel handles the stats view
type StatsModel struct {
	Character              *domain.Character
	selectedIndex          int
	editing                bool
	editingDamage          bool
	editingResource        bool
	confirmResourceDelete  bool
	resourceFormMode       resourceFormMode
	damageAction           damageAction
	editInput              textinput.Model
	damageInput            textinput.Model
	resourceNameInput      textinput.Model
	resourceTotalInput     textinput.Model
	resourceRemainingInput textinput.Model
	activeResourceInput    int
	keyMap                 StatsKeyMap
}

type damageAction int

const (
	damageActionTake damageAction = iota
	damageActionHeal
)

type resourceFormMode int

const (
	resourceFormAdd resourceFormMode = iota
	resourceFormEdit
)

type statFieldID int

const (
	fieldName statFieldID = iota
	fieldClass
	fieldFaithPath
	fieldMentor
	fieldPoise
	fieldPoiseDie
	fieldPoisePoints
	fieldMaxHP
	fieldDamage
	fieldDamageNote
	fieldSpecial
	fieldParryAndRepost
)

type statFieldEditType int

const (
	fieldEditText statFieldEditType = iota
	fieldEditNonNegativeInt
	fieldEditPositiveInt
)

type statField struct {
	id       statFieldID
	label    string
	value    string
	editType statFieldEditType
}

// NewStatsModel creates a new stats view model
func NewStatsModel(Character *domain.Character) *StatsModel {
	input := textinput.New()
	input.CharLimit = 120
	damageInput := textinput.New()
	damageInput.CharLimit = 3
	damageInput.Placeholder = "1"
	resourceNameInput := textinput.New()
	resourceNameInput.CharLimit = 50
	resourceNameInput.Placeholder = "Resource name"
	resourceTotalInput := textinput.New()
	resourceTotalInput.CharLimit = 3
	resourceTotalInput.Placeholder = "Total"
	resourceRemainingInput := textinput.New()
	resourceRemainingInput.CharLimit = 3
	resourceRemainingInput.Placeholder = "Remaining"
	return &StatsModel{
		Character:              Character,
		keyMap:                 DefaultStatsKeyMap,
		editInput:              input,
		damageInput:            damageInput,
		resourceNameInput:      resourceNameInput,
		resourceTotalInput:     resourceTotalInput,
		resourceRemainingInput: resourceRemainingInput,
	}
}

// IsInputMode returns true when the view is capturing text input
func (m *StatsModel) IsInputMode() bool {
	return m.editing || m.editingDamage || m.editingResource
}

// HandlesBeforeGlobal allows Stats-local action keys to win over app-level
// navigation when they would otherwise conflict.
func (m *StatsModel) HandlesBeforeGlobal(msg tea.KeyMsg) bool {
	return key.Matches(msg, m.keyMap.Heal)
}

// Init implements tea.Model
func (m *StatsModel) Init() tea.Cmd {
	return nil
}

// totalSelectableItems returns the total number of items the cursor can land on
func (m *StatsModel) totalSelectableItems() int {
	return len(m.statsFields()) + len(m.Character.Resources)
}

func (m *StatsModel) statsFields() []statField {
	fields := make([]statField, 0, 12)
	fields = append(fields, m.profileFields()...)
	fields = append(fields, m.combatFields()...)
	fields = append(fields, m.abilityFields()...)
	return fields
}

func (m *StatsModel) profileFields() []statField {
	return []statField{
		{fieldName, "Name", m.Character.Name, fieldEditText},
		{fieldClass, "Class", m.Character.Class, fieldEditText},
		{fieldFaithPath, "Faith Path", m.Character.FaithPath, fieldEditText},
		{fieldMentor, "Mentor", m.Character.Mentor, fieldEditText},
	}
}

func (m *StatsModel) combatFields() []statField {
	return []statField{
		{fieldPoise, "Poise", fmt.Sprintf("%d", m.Character.Stats.Poise), fieldEditNonNegativeInt},
		{fieldPoiseDie, "Poise Die", m.Character.Stats.PoiseDie, fieldEditText},
		{fieldPoisePoints, "Poise Points", fmt.Sprintf("%d", m.Character.Stats.PoisePoints), fieldEditNonNegativeInt},
		{fieldMaxHP, "Max HP", fmt.Sprintf("%d", m.Character.Stats.EffectiveMaxHP()), fieldEditPositiveInt},
		{fieldDamage, "Damage", fmt.Sprintf("%d", m.Character.Stats.DamageTaken), fieldEditNonNegativeInt},
		{fieldDamageNote, "Damage Note", m.Character.Stats.DamageNote, fieldEditText},
	}
}

func (m *StatsModel) abilityFields() []statField {
	return []statField{
		{fieldSpecial, "Special", m.Character.Stats.Special, fieldEditText},
		{fieldParryAndRepost, "Parry & Repost", m.Character.Stats.ParryAndRepost, fieldEditText},
	}
}

func (m *StatsModel) selectedField() (statField, bool) {
	fields := m.statsFields()
	if m.selectedIndex < 0 || m.selectedIndex >= len(fields) {
		return statField{}, false
	}
	return fields[m.selectedIndex], true
}

func (m *StatsModel) selectedResourceIndex() int {
	idx := m.selectedIndex - len(m.statsFields())
	if idx < 0 || idx >= len(m.Character.Resources) {
		return -1
	}
	return idx
}

func (m *StatsModel) selectedFieldID() (statFieldID, bool) {
	field, ok := m.selectedField()
	if !ok {
		return 0, false
	}
	return field.id, true
}

// Update implements tea.Model
func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editing {
			return m.updateEditMode(msg)
		}
		if m.editingDamage {
			return m.updateDamageMode(msg)
		}
		if m.editingResource {
			return m.updateResourceMode(msg)
		}
		if m.confirmResourceDelete {
			return m.updateConfirmResourceDelete(msg)
		}

		switch {
		case key.Matches(msg, m.keyMap.Up):
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case key.Matches(msg, m.keyMap.Down):
			if m.selectedIndex < m.totalSelectableItems()-1 {
				m.selectedIndex++
			}
		case key.Matches(msg, m.keyMap.Enter):
			if id, ok := m.selectedFieldID(); ok && id == fieldDamage {
				return m.enterDamageMode(damageActionTake)
			}
			return m.enterEditMode()
		case key.Matches(msg, m.keyMap.Damage):
			return m.enterDamageMode(damageActionTake)
		case key.Matches(msg, m.keyMap.Heal):
			return m.enterDamageMode(damageActionHeal)
		case key.Matches(msg, m.keyMap.Edit):
			return m.enterEditMode()
		case key.Matches(msg, m.keyMap.Add):
			return m.enterResourceMode(resourceFormAdd)
		case key.Matches(msg, m.keyMap.Delete):
			if m.selectedResourceIndex() >= 0 {
				m.confirmResourceDelete = true
				return m, nil
			}
			return m, statusCmd("Only resources can be deleted from Stats.", true)
		case key.Matches(msg, m.keyMap.Use):
			updated := m.applyDelta(-1)
			if updated {
				return m, characterUpdatedCmd(m.Character)
			}
			return m, statusCmd(m.blockedDeltaMessage(-1), true)
		case key.Matches(msg, m.keyMap.Restore):
			updated := m.applyDelta(1)
			if updated {
				return m, characterUpdatedCmd(m.Character)
			}
			return m, statusCmd(m.blockedDeltaMessage(1), true)
		case key.Matches(msg, m.keyMap.Rest):
			m.Character.Rest()
			return m, tea.Batch(
				characterUpdatedCmd(m.Character),
				statusCmd("Rested at bonfire. All resources restored.", false),
			)
		}
	}

	return m, nil
}

func (m *StatsModel) enterEditMode() (tea.Model, tea.Cmd) {
	if m.selectedResourceIndex() >= 0 {
		return m.enterResourceMode(resourceFormEdit)
	}

	field, ok := m.selectedField()
	if !ok {
		return m, nil
	}
	m.editInput.SetValue(field.value)
	m.editInput.Focus()
	m.editing = true
	return m, textinput.Blink
}

func (m *StatsModel) enterDamageMode(action damageAction) (tea.Model, tea.Cmd) {
	m.damageAction = action
	m.damageInput.SetValue("")
	m.damageInput.Focus()
	m.editingDamage = true
	return m, textinput.Blink
}

func (m *StatsModel) updateEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Escape):
		m.editing = false
		return m, nil
	case key.Matches(msg, m.keyMap.Enter):
		field, ok := m.selectedField()
		if !ok {
			m.editing = false
			return m, nil
		}
		if err := m.applyFieldValue(field, m.editInput.Value()); err != nil {
			return m, statusCmd(err.Error(), true)
		}
		m.editing = false
		return m, tea.Batch(
			characterUpdatedCmd(m.Character),
			statusCmd(fmt.Sprintf("Updated %s.", field.label), false),
		)
	}

	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m *StatsModel) applyFieldValue(field statField, value string) error {
	value = strings.TrimSpace(value)

	switch field.editType {
	case fieldEditPositiveInt:
		n, err := parsePositiveInt(value, 1)
		if err != nil {
			return fmt.Errorf("%s must be a positive number", field.label)
		}
		return m.setFieldInt(field.id, n)
	case fieldEditNonNegativeInt:
		n, err := parseNonNegativeInt(value, 0)
		if err != nil {
			return fmt.Errorf("%s must be zero or a positive number", field.label)
		}
		return m.setFieldInt(field.id, n)
	default:
		m.setFieldText(field.id, value)
		return nil
	}
}

func (m *StatsModel) setFieldInt(id statFieldID, value int) error {
	switch id {
	case fieldPoise:
		m.Character.Stats.Poise = value
	case fieldPoisePoints:
		m.Character.Stats.PoisePoints = value
	case fieldMaxHP:
		m.Character.Stats.MaxHP = value
	case fieldDamage:
		m.Character.Stats.DamageTaken = value
	default:
		return fmt.Errorf("field is not numeric")
	}
	return nil
}

func (m *StatsModel) setFieldText(id statFieldID, value string) {
	switch id {
	case fieldName:
		m.Character.Name = value
	case fieldClass:
		m.Character.Class = value
	case fieldFaithPath:
		m.Character.FaithPath = value
	case fieldMentor:
		m.Character.Mentor = value
	case fieldPoiseDie:
		m.Character.Stats.PoiseDie = value
	case fieldDamageNote:
		m.Character.Stats.DamageNote = value
	case fieldSpecial:
		m.Character.Stats.Special = value
	case fieldParryAndRepost:
		m.Character.Stats.ParryAndRepost = value
	}
}

func (m *StatsModel) updateDamageMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Escape):
		m.editingDamage = false
		return m, nil
	case key.Matches(msg, m.keyMap.Enter):
		amount, err := parsePositiveInt(m.damageInput.Value(), 1)
		if err != nil {
			return m, statusCmd("Amount must be a positive number", true)
		}

		changed, status := m.applyDamageAction(amount)
		m.editingDamage = false
		if !changed {
			return m, statusCmd(status, true)
		}
		return m, tea.Batch(
			characterUpdatedCmd(m.Character),
			statusCmd(status, false),
		)
	case msg.String() == "tab", msg.String() == "left", msg.String() == "right", msg.String() == " ":
		m.toggleDamageAction()
		return m, nil
	}

	var cmd tea.Cmd
	m.damageInput, cmd = m.damageInput.Update(msg)
	return m, cmd
}

func (m *StatsModel) enterResourceMode(mode resourceFormMode) (tea.Model, tea.Cmd) {
	m.resourceFormMode = mode
	m.activeResourceInput = 0

	if mode == resourceFormEdit {
		idx := m.selectedResourceIndex()
		if idx < 0 {
			return m, nil
		}
		resource := m.Character.Resources[idx]
		m.resourceNameInput.SetValue(resource.Name)
		m.resourceTotalInput.SetValue(fmt.Sprintf("%d", resource.Total))
		m.resourceRemainingInput.SetValue(fmt.Sprintf("%d", resource.Remaining))
	} else {
		m.resourceNameInput.SetValue("")
		m.resourceTotalInput.SetValue("")
		m.resourceRemainingInput.SetValue("")
	}

	m.focusActiveResourceInput()
	m.editingResource = true
	return m, textinput.Blink
}

func (m *StatsModel) updateResourceMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Escape):
		m.editingResource = false
		return m, nil
	case key.Matches(msg, m.keyMap.Enter):
		if m.activeResourceInput < 2 {
			m.activeResourceInput++
			m.focusActiveResourceInput()
			return m, textinput.Blink
		}
		return m.submitResource()
	case msg.String() == "tab":
		m.activeResourceInput = (m.activeResourceInput + 1) % 3
		m.focusActiveResourceInput()
		return m, textinput.Blink
	case msg.String() == "shift+tab":
		m.activeResourceInput = (m.activeResourceInput + 2) % 3
		m.focusActiveResourceInput()
		return m, textinput.Blink
	}

	var cmd tea.Cmd
	switch m.activeResourceInput {
	case 0:
		m.resourceNameInput, cmd = m.resourceNameInput.Update(msg)
	case 1:
		m.resourceTotalInput, cmd = m.resourceTotalInput.Update(msg)
	case 2:
		m.resourceRemainingInput, cmd = m.resourceRemainingInput.Update(msg)
	}
	return m, cmd
}

func (m *StatsModel) focusActiveResourceInput() {
	m.resourceNameInput.Blur()
	m.resourceTotalInput.Blur()
	m.resourceRemainingInput.Blur()

	switch m.activeResourceInput {
	case 0:
		m.resourceNameInput.Focus()
	case 1:
		m.resourceTotalInput.Focus()
	case 2:
		m.resourceRemainingInput.Focus()
	}
}

func (m *StatsModel) submitResource() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.resourceNameInput.Value())
	if name == "" {
		return m, statusCmd("Resource name is required", true)
	}

	total, err := parsePositiveInt(m.resourceTotalInput.Value(), 1)
	if err != nil {
		return m, statusCmd("Resource total must be a positive number", true)
	}

	remaining, err := parseNonNegativeInt(m.resourceRemainingInput.Value(), total)
	if err != nil {
		return m, statusCmd("Resource remaining must be zero or a positive number", true)
	}
	if remaining > total {
		return m, statusCmd("Resource remaining cannot be greater than total", true)
	}

	resource := domain.Resource{
		Name:      name,
		Total:     total,
		Remaining: remaining,
	}

	if m.resourceFormMode == resourceFormAdd {
		m.Character.Resources = append(m.Character.Resources, resource)
		m.selectedIndex = len(m.statsFields()) + len(m.Character.Resources) - 1
	} else {
		idx := m.selectedResourceIndex()
		if idx < 0 {
			m.editingResource = false
			return m, nil
		}
		m.Character.Resources[idx] = resource
	}

	m.editingResource = false
	return m, tea.Batch(
		characterUpdatedCmd(m.Character),
		statusCmd(fmt.Sprintf("Updated %s.", resource.Name), false),
	)
}

func (m *StatsModel) updateConfirmResourceDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		idx := m.selectedResourceIndex()
		if idx < 0 {
			m.confirmResourceDelete = false
			return m, nil
		}
		name := m.Character.Resources[idx].Name
		m.Character.Resources = append(m.Character.Resources[:idx], m.Character.Resources[idx+1:]...)
		if m.selectedIndex >= m.totalSelectableItems() {
			m.selectedIndex = m.totalSelectableItems() - 1
		}
		m.confirmResourceDelete = false
		return m, tea.Batch(
			characterUpdatedCmd(m.Character),
			statusCmd(fmt.Sprintf("Deleted %s.", name), false),
		)
	case "n", "N", "esc":
		m.confirmResourceDelete = false
	}
	return m, nil
}

func (m *StatsModel) toggleDamageAction() {
	if m.damageAction == damageActionTake {
		m.damageAction = damageActionHeal
		return
	}
	m.damageAction = damageActionTake
}

func (m *StatsModel) applyDamageAction(amount int) (bool, string) {
	switch m.damageAction {
	case damageActionHeal:
		before := m.Character.Stats.DamageTaken
		if before == 0 {
			return false, "HP is already full."
		}
		m.Character.Stats.DamageTaken -= amount
		if m.Character.Stats.DamageTaken < 0 {
			m.Character.Stats.DamageTaken = 0
		}
		healed := before - m.Character.Stats.DamageTaken
		return true, m.damageStatus(fmt.Sprintf("Healed %d damage.", healed))
	default:
		m.Character.Stats.DamageTaken += amount
		return true, m.damageStatus(fmt.Sprintf("Took %d damage.", amount))
	}
}

func (m *StatsModel) damageStatus(prefix string) string {
	status := m.hpStatus()
	if prefix == "" {
		return status
	}
	return fmt.Sprintf("%s %s", prefix, status)
}

func (m *StatsModel) hpStatus() string {
	maxHP := m.Character.Stats.EffectiveMaxHP()
	remaining := maxHP - m.Character.Stats.DamageTaken
	if remaining < 0 {
		remaining = 0
	}
	return fmt.Sprintf("HP %d/%d.", remaining, maxHP)
}

// applyDelta modifies the currently selected stat or resource by delta
func (m *StatsModel) applyDelta(delta int) bool {
	if field, ok := m.selectedField(); ok {
		var current *int
		actualDelta := delta
		min := 0
		switch field.id {
		case fieldPoise:
			current = &m.Character.Stats.Poise
		case fieldPoisePoints:
			current = &m.Character.Stats.PoisePoints
		case fieldMaxHP:
			current = &m.Character.Stats.MaxHP
			min = 1
			if *current <= 0 {
				*current = m.Character.Stats.EffectiveMaxHP()
			}
		case fieldDamage:
			current = &m.Character.Stats.DamageTaken
			actualDelta = -delta
		default:
			return false
		}

		next := *current + actualDelta
		if next < min {
			next = min
		}
		if next == *current {
			return false
		}
		*current = next
		return true
	}

	// Editing a resource
	resIdx := m.selectedResourceIndex()
	if resIdx >= 0 && resIdx < len(m.Character.Resources) {
		if delta < 0 {
			return m.Character.Resources[resIdx].Use()
		}
		return m.Character.Resources[resIdx].Restore()
	}
	return false
}

func (m *StatsModel) blockedDeltaMessage(delta int) string {
	if field, ok := m.selectedField(); ok {
		switch field.id {
		case fieldPoise:
			return "Poise cannot go below 0."
		case fieldPoisePoints:
			return "Poise Points cannot go below 0."
		case fieldMaxHP:
			return "Max HP cannot go below 1."
		case fieldDamage:
			if delta > 0 {
				return "Damage is already 0."
			}
			return "Damage was not changed."
		default:
			return fmt.Sprintf("%s is not adjustable with +/-.", field.label)
		}
	}

	resIdx := m.selectedResourceIndex()
	if resIdx >= 0 && resIdx < len(m.Character.Resources) {
		resource := m.Character.Resources[resIdx]
		if delta < 0 {
			return fmt.Sprintf("%s is empty.", resource.Name)
		}
		return fmt.Sprintf("%s is already full.", resource.Name)
	}

	return "Nothing selected."
}

// wideLabel is a label style wide enough for "Parry & Repost"
var wideLabel = styles.LabelStyle.Width(16)

// View implements tea.Model
func (m *StatsModel) View() string {
	var b strings.Builder

	if m.editingDamage {
		return m.viewDamageForm()
	}
	if m.editingResource {
		return m.viewResourceForm()
	}
	if m.confirmResourceDelete {
		return m.viewConfirmResourceDelete()
	}

	b.WriteString(styles.SectionStyle.Render("PROFILE") + "\n\n")
	profileFields := m.profileFields()
	m.renderFieldRows(&b, 0, profileFields)

	b.WriteString(styles.SectionStyle.Render("COMBAT") + "\n\n")
	combatStart := len(profileFields)
	combatFields := m.combatFields()
	m.renderFieldRows(&b, combatStart, combatFields)

	// HP bar beneath combat stats
	maxHP := m.Character.Stats.EffectiveMaxHP()
	hpBar := styles.RenderHPBar(m.Character.Stats.DamageTaken, maxHP)
	remaining := maxHP - m.Character.Stats.DamageTaken
	if remaining < 0 {
		remaining = 0
	}
	b.WriteString(fmt.Sprintf("\n  %s %s  %s\n",
		wideLabel.Render("HP"),
		hpBar,
		styles.DimmedStyle.Render(fmt.Sprintf("%d/%d", remaining, maxHP)),
	))

	b.WriteString("\n")
	b.WriteString(styles.SectionStyle.Render("ABILITIES") + "\n\n")
	abilityStart := combatStart + len(combatFields)
	m.renderFieldRows(&b, abilityStart, m.abilityFields())

	// Resources section
	b.WriteString("\n")
	b.WriteString(styles.SectionStyle.Render("RESOURCES") + "\n\n")

	if len(m.Character.Resources) == 0 {
		b.WriteString(styles.EmptyStateStyle.Render("No resources.") + "\n")
	}

	for i, res := range m.Character.Resources {
		cursor := "  "
		nameStyle := styles.NormalStyle
		if i+len(m.statsFields()) == m.selectedIndex {
			cursor = "> "
			nameStyle = styles.SelectedStyle
		}

		dots := styles.RenderResourceDots(res.Remaining, res.Total)
		name := nameStyle.Render(fmt.Sprintf("%-16s", res.Name))
		count := styles.DimmedStyle.Render(fmt.Sprintf("(%d/%d)", res.Remaining, res.Total))

		b.WriteString(fmt.Sprintf("%s%s  %s %s\n", cursor, name, dots, count))
	}

	// Help for this view
	b.WriteString("\n")
	help := []string{
		styles.KeyStyle.Render("[j/k]") + " select",
		styles.KeyStyle.Render("[enter/e]") + " edit",
		styles.KeyStyle.Render("[a]") + " add resource",
		styles.KeyStyle.Render("[r]") + " rest",
	}
	if id, ok := m.selectedFieldID(); ok && id == fieldDamage {
		help = append(help,
			styles.KeyStyle.Render("[enter/d]")+" damage",
			styles.KeyStyle.Render("[h]")+" heal",
		)
	} else if m.selectedResourceIndex() >= 0 {
		help = append(help,
			styles.KeyStyle.Render("[-]")+" use",
			styles.KeyStyle.Render("[+/=]")+" restore",
			styles.KeyStyle.Render("[x]")+" delete",
		)
	} else {
		help = append(help, styles.KeyStyle.Render("[+/-]")+" adjust")
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}

func (m *StatsModel) renderFieldRows(b *strings.Builder, start int, fields []statField) {
	for i, field := range fields {
		index := start + i
		cursor := "  "
		labelStyle := wideLabel
		if index == m.selectedIndex {
			cursor = "> "
			labelStyle = styles.SelectedStyle.Width(16)
		}

		label := labelStyle.Render(field.label)
		if m.editing && index == m.selectedIndex {
			b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, m.editInput.View()))
			continue
		}

		value := styles.ValueStyle.Render(field.value)
		if field.value == "" {
			value = styles.DimmedStyle.Render("(empty)")
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, value))
	}
	b.WriteString("\n")
}

func (m *StatsModel) viewDamageForm() string {
	var b strings.Builder

	b.WriteString(styles.SectionStyle.Render("DAMAGE") + "\n\n")

	takeStyle := styles.NormalStyle
	healStyle := styles.NormalStyle
	if m.damageAction == damageActionTake {
		takeStyle = styles.SelectedStyle
	} else {
		healStyle = styles.SelectedStyle
	}

	b.WriteString("  " + takeStyle.Render("Take Damage") + "  ")
	b.WriteString(styles.DimmedStyle.Render("/") + "  ")
	b.WriteString(healStyle.Render("Heal") + "\n\n")

	b.WriteString("  " + styles.LabelStyle.Render("Amount") + " " + m.damageInput.View() + "\n\n")
	b.WriteString("  " + styles.LabelStyle.Render("Current HP") + " " + styles.ValueStyle.Render(m.damageStatus("")) + "\n\n")

	help := []string{
		styles.KeyStyle.Render("[tab/←/→]") + " switch",
		styles.KeyStyle.Render("[enter]") + " apply",
		styles.KeyStyle.Render("[esc]") + " cancel",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}

func (m *StatsModel) viewResourceForm() string {
	var b strings.Builder

	title := "ADD RESOURCE"
	if m.resourceFormMode == resourceFormEdit {
		title = "EDIT RESOURCE"
	}
	b.WriteString(styles.SectionStyle.Render(title) + "\n\n")

	inputs := []struct {
		label string
		input textinput.Model
	}{
		{"Name", m.resourceNameInput},
		{"Total", m.resourceTotalInput},
		{"Remaining", m.resourceRemainingInput},
	}

	for i, input := range inputs {
		style := styles.NormalStyle
		if i == m.activeResourceInput {
			style = styles.SelectedStyle
		}
		b.WriteString(fmt.Sprintf("  %s\n", style.Render(input.label)))
		b.WriteString(fmt.Sprintf("  %s\n\n", input.input.View()))
	}

	help := []string{
		styles.KeyStyle.Render("[tab]") + " next field",
		styles.KeyStyle.Render("[enter]") + " submit",
		styles.KeyStyle.Render("[esc]") + " cancel",
	}
	b.WriteString(styles.HelpStyle.Render(strings.Join(help, "  ")))

	return b.String()
}

func (m *StatsModel) viewConfirmResourceDelete() string {
	var b strings.Builder

	idx := m.selectedResourceIndex()
	if idx < 0 {
		return ""
	}

	resource := m.Character.Resources[idx]
	b.WriteString(styles.SectionStyle.Render("CONFIRM DELETE") + "\n\n")
	b.WriteString(fmt.Sprintf("  Delete %s?\n\n", styles.SelectedStyle.Render(resource.Name)))
	b.WriteString("  " + styles.KeyStyle.Render("[y]") + " yes  ")
	b.WriteString(styles.KeyStyle.Render("[n]") + " no\n")

	return b.String()
}
