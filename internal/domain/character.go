package domain

// Character represents a Sword & Board TTRPG character sheet
type Character struct {
	Name      string     `yaml:"name"`
	Class     string     `yaml:"class"`
	FaithPath string     `yaml:"faith_path"`
	Mentor    string     `yaml:"mentor"`
	Stats     Stats      `yaml:"stats"`
	Resources []Resource `yaml:"resources"`
	Inventory []Item     `yaml:"inventory"`
	Spells    []Spell    `yaml:"spells"`
}

// Stats holds the character's core statistics
type Stats struct {
	Poise          int    `yaml:"poise"`
	PoiseDie       string `yaml:"poise_die"`
	PoisePoints    int    `yaml:"poise_points"`
	DamageTaken    int    `yaml:"damage_taken"`
	DamageNote     string `yaml:"damage_note"`
	Special        string `yaml:"special"`
	ParryAndRepost string `yaml:"parry_and_repost,omitempty"`
}

// Resource represents a consumable resource like Estus Flask
type Resource struct {
	Name      string `yaml:"name"`
	Total     int    `yaml:"total"`
	Remaining int    `yaml:"remaining"`
}

// Use decrements the remaining count of a resource
func (r *Resource) Use() bool {
	if r.Remaining > 0 {
		r.Remaining--
		return true
	}
	return false
}

// Restore increments the remaining count up to total
func (r *Resource) Restore() bool {
	if r.Remaining < r.Total {
		r.Remaining++
		return true
	}
	return false
}

// RestoreFull resets remaining to total
func (r *Resource) RestoreFull() {
	r.Remaining = r.Total
}

// ItemType categorizes inventory items
type ItemType string

const (
	ItemTypeWeapon     ItemType = "Weapon"
	ItemTypeEquipment  ItemType = "Equipment"
	ItemTypeConsumable ItemType = "Consumable"
)

// Item represents an inventory item
type Item struct {
	Name     string   `yaml:"name"`
	Quantity int      `yaml:"quantity"`
	Type     ItemType `yaml:"type"`
	Notes    string   `yaml:"notes,omitempty"`
}

// Spell represents a spell or miracle the character can cast
type Spell struct {
	Name      string `yaml:"name"`
	TotalUses int    `yaml:"total_uses"`
	Used      int    `yaml:"used"`
}

// Use marks one use of the spell
func (s *Spell) Use() bool {
	if s.Used < s.TotalUses {
		s.Used++
		return true
	}
	return false
}

// Restore recovers one use of the spell
func (s *Spell) Restore() bool {
	if s.Used > 0 {
		s.Used--
		return true
	}
	return false
}

// RestoreFull resets used count to zero
func (s *Spell) RestoreFull() {
	s.Used = 0
}

// Remaining returns how many uses are left
func (s *Spell) Remaining() int {
	return s.TotalUses - s.Used
}

// Rest restores all resources and spells to full
func (c *Character) Rest() {
	for i := range c.Resources {
		c.Resources[i].RestoreFull()
	}
	for i := range c.Spells {
		c.Spells[i].RestoreFull()
	}
}

// ItemsByType groups inventory items by their type
func (c *Character) ItemsByType() map[ItemType][]Item {
	result := make(map[ItemType][]Item)
	for _, item := range c.Inventory {
		result[item.Type] = append(result[item.Type], item)
	}
	return result
}

// AddItem adds a new item to the inventory
func (c *Character) AddItem(item Item) {
	c.Inventory = append(c.Inventory, item)
}

// RemoveItem removes an item at the given index
func (c *Character) RemoveItem(index int) bool {
	if index < 0 || index >= len(c.Inventory) {
		return false
	}
	c.Inventory = append(c.Inventory[:index], c.Inventory[index+1:]...)
	return true
}

// AddSpell adds a new spell
func (c *Character) AddSpell(spell Spell) {
	c.Spells = append(c.Spells, spell)
}

// RemoveSpell removes a spell at the given index
func (c *Character) RemoveSpell(index int) bool {
	if index < 0 || index >= len(c.Spells) {
		return false
	}
	c.Spells = append(c.Spells[:index], c.Spells[index+1:]...)
	return true
}
