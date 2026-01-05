package game

import (
	"math/rand"

	"github.com/Bloodisck/bootdev-pokedex/internal/pokeapi"
)

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

type EvolutionRequirement struct {
	NextStage     string // The species name to evolve into
	RequiredLevel int    // Minimum level
	RequiredStone string // The specific stone name (e.g., "fire-stone")
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

// In internal/game/models.go

// Update signature to accept Client
func NewBattlePokemon(base pokeapi.Pokemon, level int, client pokeapi.Client) (*BattlePokemon, error) {
	bp := &BattlePokemon{
		Base:     base,
		Nickname: base.Name,
		Level:    level,
		XP:       0,
		Status:   StatusNone,
		Stats:    Stats{},
	}

	bp.NextLevelXP = level * level * 10
	bp.RecalculateStats()
	bp.Stats.HP = bp.Stats.MaxHP

	// 1. Pick a random move from the list of possible moves
	if len(base.Moves) > 0 {
		// Simple logic: Pick a random one.
		randIdx := rand.Intn(len(base.Moves))
		moveName := base.Moves[randIdx].Move.Name

		// 2. Fetch the details
		apiMove, err := client.GetMove(moveName)
		if err == nil {
			// Convert API Move to Game Move
			gameMove := Move{
				Name:      apiMove.Name,
				Type:      apiMove.Type.Name,
				Power:     apiMove.Power,
				Accuracy:  apiMove.Accuracy,
				MaxPP:     apiMove.PP,
				CurrentPP: apiMove.PP,
			}
			bp.Moves = append(bp.Moves, gameMove)
		}
	}

	// Fallback if API failed or no moves found
	if len(bp.Moves) == 0 {
		bp.Moves = []Move{{Name: "Tackle", Type: "normal", Power: 40, MaxPP: 35, CurrentPP: 35}}
	}

	return bp, nil
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

func (p *BattlePokemon) Evolve(newBase pokeapi.Pokemon, client pokeapi.Client) {
	oldMaxHP := p.Stats.MaxHP
	oldSpeciesName := p.Base.Name // Store the old name (e.g., "charmander")

	// 1. Update Identity
	// We check against the OLD name to see if it was default
	if p.Nickname == oldSpeciesName {
		p.Nickname = newBase.Name
	}

	// NOW we update the base to the new species
	p.Base = newBase

	// 2. Recalculate Stats
	p.RecalculateStats()

	// 3. Scale HP proportionally
	hpPercent := float64(p.Stats.HP) / float64(oldMaxHP)
	p.Stats.HP = int(hpPercent * float64(p.Stats.MaxHP))

	if p.Stats.HP < 1 && p.Status != StatusFainted {
		p.Stats.HP = 1
	}

	// 4. Learn a New Move
	if len(newBase.Moves) > 0 {
		randomIndex := rand.Intn(len(newBase.Moves))
		moveName := newBase.Moves[randomIndex].Move.Name
		apiMove, err := client.GetMove(moveName)
		if err == nil {
			newMove := Move{
				Name:      apiMove.Name,
				Type:      apiMove.Type.Name,
				Power:     apiMove.Power,
				Accuracy:  apiMove.Accuracy,
				MaxPP:     apiMove.PP,
				CurrentPP: apiMove.PP,
			}

			hasMove := false
			for _, m := range p.Moves {
				if m.Name == newMove.Name {
					hasMove = true
				}
			}
			if !hasMove {
				if len(p.Moves) >= 4 {
					p.Moves[0] = newMove
				} else {
					p.Moves = append(p.Moves, newMove)
				}
			}
		}
	}
}
