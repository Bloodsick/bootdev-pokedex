package game

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

// StartBattle is the main entry point
func StartBattle(party []*BattlePokemon, wildPokemon *BattlePokemon, inventory *PlayerInventory) bool {
	scanner := bufio.NewScanner(os.Stdin)

	// Get first alive pokemon
	var activeMon *BattlePokemon
	for _, p := range party {
		if p.Status != StatusFainted && p.Stats.HP > 0 {
			activeMon = p
			break
		}
	}

	if activeMon == nil {
		fmt.Println("Your entire team is fainted! You assume the fetal position and cry.")
		return false
	}

	fmt.Printf("\n--- BATTLE STARTED: %s vs Wild %s ---\n", activeMon.Nickname, wildPokemon.Nickname)

	for {
		// --- 1. Pre-Turn Status Check (Burn/Poison damage could go here, usually goes end of turn) ---

		// --- 2. Player Input ---
		fmt.Printf("\n%s (Lvl %d): %d/%d HP\n", activeMon.Nickname, activeMon.Level, activeMon.Stats.HP, activeMon.Stats.MaxHP)
		fmt.Printf("Wild %s (Lvl %d): %d/%d HP\n", wildPokemon.Nickname, wildPokemon.Level, wildPokemon.Stats.HP, wildPokemon.Stats.MaxHP)

		fmt.Println("Choose: (1) Fight  (2) Bag  (3) Pokemon  (4) Run")
		fmt.Print("> ")
		scanner.Scan()
		choice := scanner.Text()

		turnEnded := false

		switch choice {
		case "1": // FIGHT
			turnEnded = handleFightMenu(scanner, activeMon, wildPokemon)

		case "2": // BAG (Catching happens here)
			caught, usedTurn := handleBagMenu(scanner, inventory, wildPokemon, activeMon)
			if caught {
				return true // Battle ends, pokemon caught
			}
			turnEnded = usedTurn

		case "3": // POKEMON (Switching)
			newMon := handleSwitchMenu(scanner, party)
			if newMon != nil {
				activeMon = newMon
				fmt.Printf("Go! %s!\n", activeMon.Nickname)
				turnEnded = true
			}

		case "4": // RUN
			// Run formula: Speed check
			if activeMon.Stats.Speed >= wildPokemon.Stats.Speed || rand.Intn(100) < 50 {
				fmt.Println("Got away safely!")
				return false
			}
			fmt.Println("Can't escape!")
			turnEnded = true
		}

		if !turnEnded {
			continue
		}

		// --- 3. Enemy Turn ---
		if wildPokemon.Stats.HP > 0 {
			// Simple AI: Random move
			move := wildPokemon.Moves[rand.Intn(len(wildPokemon.Moves))]
			performMove(wildPokemon, activeMon, move)

			if activeMon.Stats.HP <= 0 {
				activeMon.Stats.HP = 0
				activeMon.Status = StatusFainted
				fmt.Printf("%s fainted!\n", activeMon.Nickname)

				// Force switch or lose
				if !forceSwitch(scanner, party, &activeMon) {
					fmt.Println("You blacked out...")
					return false
				}
			}
		} else {
			// Enemy Fainted
			fmt.Printf("Wild %s fainted!\n", wildPokemon.Nickname)

			// Calculate reward based on level
			goldReward := wildPokemon.Level*50 + rand.Intn(30)
			inventory.Money += goldReward
			fmt.Printf("You received ₽%d for winning!\n", goldReward)

			distributeXP(activeMon, wildPokemon)
			return true // Win
		}
	}
}

func performMove(attacker, defender *BattlePokemon, move Move) {
	fmt.Printf("%s used %s!\n", attacker.Nickname, move.Name)

	// Accuracy Check
	if rand.Intn(100) > move.Accuracy {
		fmt.Println("...but it missed!")
		return
	}

	// Damage Calc
	// ((2 * Level / 5 + 2) * Power * A / D) / 50 + 2
	atk := float64(attacker.Stats.Attack)
	def := float64(defender.Stats.Defense)
	damage := (((2.0*float64(attacker.Level)/5.0 + 2.0) * float64(move.Power) * atk / def) / 50.0) + 2.0

	// Apply Type Effectiveness (Simplified)
	// You should integrate the 'types.go' logic from previous step here

	finalDamage := int(damage)
	if finalDamage < 1 {
		finalDamage = 1
	}

	defender.Stats.HP -= finalDamage
	if defender.Stats.HP < 0 {
		defender.Stats.HP = 0
	}
	fmt.Printf("Dealt %d damage.\n", finalDamage)

	// Apply Status
	if move.StatusEffect != StatusNone && defender.Status == StatusNone && rand.Intn(100) < 30 {
		defender.Status = move.StatusEffect
		fmt.Printf("%s was afflicted with status %d!\n", defender.Nickname, move.StatusEffect)
	}
}

func handleBagMenu(scanner *bufio.Scanner, inv *PlayerInventory, wild *BattlePokemon, playerMon *BattlePokemon) (bool, bool) {
	fmt.Println("\n--- BAG ---")
	fmt.Println("1. Pokeballs")
	fmt.Println("2. Healing Items")
	fmt.Println("3. Cancel")

	fmt.Print("Select category > ")
	scanner.Scan()
	category := scanner.Text()

	switch category {
	case "1": // POKEBALL SUB-MENU
		fmt.Println("\n--- POKEBALLS ---")
		fmt.Printf("1. Pokeball (x%d)\n", inv.Pokeballs)
		fmt.Printf("2. Greatball (x%d)\n", inv.Greatballs)
		fmt.Printf("3. Ultraball (x%d)\n", inv.Ultraballs)
		fmt.Println("4. Back")

		fmt.Print("Select ball > ")
		scanner.Scan()
		ballChoice := scanner.Text()

		var multiplier float64
		var ballName string
		count := 0

		switch ballChoice {
		case "1":
			multiplier, ballName, count = 1.0, "Pokeball", inv.Pokeballs
		case "2":
			multiplier, ballName, count = 1.5, "Greatball", inv.Greatballs
		case "3":
			multiplier, ballName, count = 2.0, "Ultraball", inv.Ultraballs
		default:
			return false, false // Go back to main battle menu
		}

		if count <= 0 {
			fmt.Printf("You don't have any %ss!\n", ballName)
			return false, false
		}

		// Deduct item
		if ballChoice == "1" {
			inv.Pokeballs--
		} else if ballChoice == "2" {
			inv.Greatballs--
		} else {
			inv.Ultraballs--
		}

		fmt.Printf("You threw a %s!\n", ballName)

		// Catch Formula with Multiplier
		chance := float64(((3*wild.Stats.MaxHP)-(2*wild.Stats.HP))*100) / float64(3*wild.Stats.MaxHP)
		if wild.Status != StatusNone {
			chance += 10
		}

		finalChance := chance * multiplier

		if rand.Intn(100) < int(finalChance) {
			fmt.Printf("Gotcha! The %s was caught!\n", wild.Base.Name)
			return true, true
		}
		fmt.Println("It broke free!")
		return false, true

	case "2": // HEALING SUB-MENU
		fmt.Println("\n--- HEALING ---")
		fmt.Printf("1. Potion (x%d) [+20 HP]\n", inv.Potions)
		fmt.Printf("2. Super Potion (x%d) [+50 HP]\n", inv.SuperPotions)
		fmt.Printf("3. Revive (x%d) [Restores Fainted]\n", inv.Revives)
		fmt.Println("4. Back")

		fmt.Print("Select item > ")
		scanner.Scan()
		itemChoice := scanner.Text()

		switch itemChoice {
		case "1", "2":
			healAmt := 20
			itemPtr := &inv.Potions
			if itemChoice == "2" {
				healAmt = 50
				itemPtr = &inv.SuperPotions
			}

			if *itemPtr <= 0 {
				fmt.Println("You don't have any!")
				return false, false
			}
			if playerMon.Status == StatusFainted {
				fmt.Println("Potions don't work on fainted Pokemon! Use a Revive.")
				return false, false
			}

			*itemPtr--
			playerMon.Stats.HP += healAmt
			if playerMon.Stats.HP > playerMon.Stats.MaxHP {
				playerMon.Stats.HP = playerMon.Stats.MaxHP
			}
			fmt.Printf("Used Potion! %s's HP is now %d/%d\n", playerMon.Nickname, playerMon.Stats.HP, playerMon.Stats.MaxHP)
			return false, true

		case "3": // REVIVE LOGIC
			if inv.Revives <= 0 {
				fmt.Println("You don't have any Revives!")
				return false, false
			}
			if playerMon.Status != StatusFainted {
				fmt.Println("That Pokemon is already conscious!")
				return false, false
			}

			inv.Revives--
			playerMon.Status = StatusNone
			playerMon.Stats.HP = playerMon.Stats.MaxHP / 2
			fmt.Printf("%s was revived to half HP!\n", playerMon.Nickname)
			return false, true
		}
	}

	return false, false
}

func distributeXP(winner, loser *BattlePokemon) {
	xpGain := (loser.Base.BaseExperience * loser.Level) / 7
	winner.XP += xpGain
	fmt.Printf("%s gained %d XP!\n", winner.Nickname, xpGain)

	if winner.XP >= winner.NextLevelXP {
		winner.Level++
		winner.XP -= winner.NextLevelXP
		winner.NextLevelXP = winner.Level * winner.Level * 10
		winner.RecalculateStats()
		fmt.Printf("%s grew to Level %d!\n", winner.Nickname, winner.Level)

		// Simple Evolution Check (Example)
		if winner.Level >= 16 && winner.Base.Name == "charmander" {
			fmt.Println("What? Charmander is evolving!")
			winner.Nickname = "Charmeleon" // In real app, fetch new base data
			// winner.Base = FetchPokemon("charmeleon")
		}
	}
}

// Helpers for menus (Switching, Fight) excluded for brevity but follow similar pattern
func handleFightMenu(scanner *bufio.Scanner, active *BattlePokemon, enemy *BattlePokemon) bool {
	for i, m := range active.Moves {
		fmt.Printf("%d. %s (%s) [%d/%d PP]\n", i+1, m.Name, m.Type, m.CurrentPP, m.MaxPP)
	}
	scanner.Scan()
	idx, _ := strconv.Atoi(scanner.Text())
	if idx > 0 && idx <= len(active.Moves) {
		performMove(active, enemy, active.Moves[idx-1])
		return true
	}
	return false
}

func forceSwitch(scanner *bufio.Scanner, party []*BattlePokemon, active **BattlePokemon) bool {
	// Loop through party to find replacement
	for _, p := range party {
		if p.Stats.HP > 0 {
			*active = p
			fmt.Printf("Go! %s!\n", p.Nickname)
			return true
		}
	}
	return false
}

func handleSwitchMenu(scanner *bufio.Scanner, party []*BattlePokemon) *BattlePokemon {
	fmt.Println("Select Pokemon:")
	for i, p := range party {
		fmt.Printf("%d. %s (%d/%d HP)\n", i+1, p.Nickname, p.Stats.HP, p.Stats.MaxHP)
	}
	scanner.Scan()
	idx, _ := strconv.Atoi(scanner.Text())
	if idx > 0 && idx <= len(party) && party[idx-1].Stats.HP > 0 {
		return party[idx-1]
	}
	return nil
}
