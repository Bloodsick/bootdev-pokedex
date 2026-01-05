package game

import "github.com/Bloodisck/bootdev-pokedex/internal/pokeapi"

// Status constants
type StatusID int

const (
	StatusNone StatusID = iota
	StatusBurn
	StatusPoison
	StatusParalysis
	StatusFainted
)

// Stats structure
type Stats struct {
	HP      int
	Attack  int
	Defense int
	Speed   int
	MaxHP   int
}

// Move represents a usable attack
type Move struct {
	Name         string
	Type         string // "fire", "water", etc.
	Power        int    // 0 for status moves
	Accuracy     int    // 0-100
	StatusEffect StatusID
	MaxPP        int
	CurrentPP    int
}

// BattlePokemon is a dynamic instance of a Pokemon
type BattlePokemon struct {
	Base        pokeapi.Pokemon
	Nickname    string
	Level       int
	XP          int
	NextLevelXP int
	Stats       Stats
	Status      StatusID
	Moves       []Move
	Nature      string // e.g., "Adamant" (+Atk, -SpAtk)
}

// PlayerInventory holds items
type PlayerInventory struct {
	Money           int
	Potions         int
	SuperPotions    int
	Pokeballs       int
	Greatballs      int
	Ultraballs      int
	Revives         int
	EvolutionStones map[string]int
}

// NewBattlePokemon creates a fresh pokemon and calculates initial stats
func NewBattlePokemon(base pokeapi.Pokemon, level int) *BattlePokemon {
	bp := &BattlePokemon{
		Base:     base,
		Nickname: base.Name,
		Level:    level,
		XP:       0,
		Status:   StatusNone,
		Nature:   "Neutral", // Simplified for now
	}

	bp.NextLevelXP = level * level * 10 // Simple parabolic curve
	bp.GenerateMoves()
	bp.RecalculateStats()
	bp.Stats.HP = bp.Stats.MaxHP // Full heal on creation
	return bp
}

func (p *BattlePokemon) RecalculateStats() {
	// Formula: ((Base * 2 * Level) / 100) + 5
	// HP Formula: ((Base * 2 * Level) / 100) + Level + 10

	baseStats := make(map[string]int)
	for _, s := range p.Base.Stats {
		baseStats[s.Stat.Name] = s.BaseStat
	}

	p.Stats.MaxHP = ((baseStats["hp"] * 2 * p.Level) / 100) + p.Level + 10
	p.Stats.Attack = ((baseStats["attack"] * 2 * p.Level) / 100) + 5
	p.Stats.Defense = ((baseStats["defense"] * 2 * p.Level) / 100) + 5
	p.Stats.Speed = ((baseStats["speed"] * 2 * p.Level) / 100) + 5
}

// GenerateMoves populates moves based on type (Simplified logic)
func (p *BattlePokemon) GenerateMoves() {
	p.Moves = []Move{
		{Name: "Tackle", Type: "normal", Power: 40, Accuracy: 100, MaxPP: 35, CurrentPP: 35},
	}

	// Add type-specific moves
	for _, t := range p.Base.Types {
		switch t.Type.Name {
		case "fire":
			p.Moves = append(p.Moves, Move{Name: "Ember", Type: "fire", Power: 40, Accuracy: 100, StatusEffect: StatusBurn, MaxPP: 25, CurrentPP: 25})
		case "water":
			p.Moves = append(p.Moves, Move{Name: "Water Gun", Type: "water", Power: 40, Accuracy: 100, MaxPP: 25, CurrentPP: 25})
		case "electric":
			p.Moves = append(p.Moves, Move{Name: "ThunderShock", Type: "electric", Power: 40, Accuracy: 100, StatusEffect: StatusParalysis, MaxPP: 30, CurrentPP: 30})
		case "poison":
			p.Moves = append(p.Moves, Move{Name: "Poison Sting", Type: "poison", Power: 15, Accuracy: 100, StatusEffect: StatusPoison, MaxPP: 35, CurrentPP: 35})
		}
	}
}

func (p *BattlePokemon) HealFull() {
	p.Stats.HP = p.Stats.MaxHP
	p.Status = StatusNone
	for i := range p.Moves {
		p.Moves[i].CurrentPP = p.Moves[i].MaxPP
	}
}

func (s StatusID) String() string {
	switch s {
	case StatusNone:
		return "OK"
	case StatusBurn:
		return "BRN"
	case StatusPoison:
		return "PSN"
	case StatusParalysis:
		return "PAR"
	case StatusFainted:
		return "FNT"
	default:
		return "???"
	}
}
