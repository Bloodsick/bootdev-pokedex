package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"

	"github.com/Bloodisck/bootdev-pokedex/internal/game"
	pokeapi "github.com/Bloodisck/bootdev-pokedex/internal/pokeapi"
)

type Config struct {
	NextURL       *string
	PreviousURL   *string
	Pokeapi       pokeapi.Client
	CaughtPokemon map[string]pokeapi.Pokemon
	VisibleAreas  []string
	Party         []*game.BattlePokemon
	PC            []*game.BattlePokemon
	Inventory     game.PlayerInventory
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
}

func main() {
	pokeClient := pokeapi.NewClient(5*time.Second, 5*time.Minute)

	cfg := &Config{
		Pokeapi: pokeClient,
	}

	loadGame(cfg)

	startRepl(cfg)
}

const saveFilePath = "savegame.json"

func saveGame(cfg *Config) error {
	// We create a temporary struct to hold EVERYTHING we want to save
	type SaveData struct {
		CaughtPokemon map[string]pokeapi.Pokemon `json:"caught_pokemon"`
		Party         []*game.BattlePokemon      `json:"party"`
		PC            []*game.BattlePokemon      `json:"pc"`
		Inventory     game.PlayerInventory       `json:"inventory"`
	}

	data := SaveData{
		CaughtPokemon: cfg.CaughtPokemon,
		Party:         cfg.Party,
		PC:            cfg.PC,
		Inventory:     cfg.Inventory,
	}

	fileData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(saveFilePath, fileData, 0644)
}

func loadGame(cfg *Config) {
	// Initialize defaults
	cfg.CaughtPokemon = make(map[string]pokeapi.Pokemon)
	cfg.Party = []*game.BattlePokemon{}
	if cfg.Inventory.EvolutionStones == nil {
		cfg.Inventory.EvolutionStones = make(map[string]int)
	}
	cfg.Inventory = game.PlayerInventory{
		Pokeballs: 20,
		Potions:   10,
	}

	fileData, err := os.ReadFile(saveFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// NEW GAME SEQUENCE
			runNewGameSequence(cfg)
			return
		}
		return
	}

	// 3. Decode the JSON
	type SaveData struct {
		CaughtPokemon map[string]pokeapi.Pokemon `json:"caught_pokemon"`
		Party         []*game.BattlePokemon      `json:"party"`
		PC            []*game.BattlePokemon      `json:"pc"`
		Inventory     game.PlayerInventory       `json:"inventory"`
	}

	var loadedData SaveData
	err = json.Unmarshal(fileData, &loadedData)
	if err != nil {
		fmt.Println("Error loading save file (it might be corrupted or old). Starting new game.")
		return
	}

	// 4. Load data into config
	cfg.CaughtPokemon = loadedData.CaughtPokemon
	cfg.Party = loadedData.Party
	cfg.PC = loadedData.PC
	cfg.Inventory = loadedData.Inventory
}

func runNewGameSequence(cfg *Config) {
	reader := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to the world of Pokémon!")
	fmt.Println("It looks like you're new here. Please choose your starter:")
	fmt.Println("1. Charmander")
	fmt.Println("2. Bulbasaur")
	fmt.Println("3. Squirtle")

	starters := map[string]string{
		"1": "charmander",
		"2": "bulbasaur",
		"3": "squirtle",
	}

	var choice string
	for {
		fmt.Print("Enter number (1-3): ")
		reader.Scan()
		input := reader.Text()

		if name, ok := starters[input]; ok {
			choice = name
			break
		}
		fmt.Println("Invalid choice. Please pick 1, 2, or 3.")
	}

	fmt.Printf("Fetching data for %s...\n", choice)

	// Fetch the base data from API
	pokemonBase, err := cfg.Pokeapi.GetPokemon(choice)
	if err != nil {
		fmt.Printf("Error starting game: %v\n", err)
		os.Exit(1)
	}

	// Create the Battle instance (Starting at Level 5)
	starter, err := game.NewBattlePokemon(pokemonBase, 5, cfg.Pokeapi)
	if err != nil {
		fmt.Printf("Error generating pokemon: %v\n", err)
	}

	cfg.Party = append(cfg.Party, starter)
	cfg.CaughtPokemon[choice] = pokemonBase // Add to Pokedex too

	fmt.Printf("\nGreat choice! %s has joined your team.\n", choice)
	fmt.Println("Use the 'map' command to find areas and 'catch' to start a battle!")

	// Initial save so they don't have to pick again if they crash
	saveGame(cfg)
}

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	pokemon := strings.Fields(text)
	return pokemon
}

func startRepl(cfg *Config) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      "Pokedex > ",
		HistoryFile: "/tmp/pokedex_history.tmp",
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}

		words := cleanInput(line)
		if len(words) == 0 {
			continue
		}

		commandName := words[0]
		args := words[1:]

		command, ok := getCommands()[commandName]
		if ok {
			err := command.callback(cfg, args)
			if err != nil {
				fmt.Println(err)
			}
			continue
		} else {
			fmt.Println("Unknown command")
			continue
		}
	}
}

func commandExit(config *Config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(cfg *Config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")

	order := []string{
		"help",
		"map",
		"mapb",
		"left",
		"right",
		"explore",
		"bag",
		"heal",
		"evolve",
		"shop",
		"catch",
		"battle",
		"team",
		"addteam",
		"inspect",
		"pokedex",
		"exit",
	}

	availableCommands := getCommands()

	for _, name := range order {
		if cmd, ok := availableCommands[name]; ok {
			fmt.Printf("%s: %s\n", cmd.name, cmd.description)
		}
	}

	return nil
}

func commandExplore(config *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: explore <name or number>")
	}

	target := args[0]

	if index, err := strconv.Atoi(target); err == nil {
		realIndex := index - 1
		if realIndex >= 0 && realIndex < len(config.VisibleAreas) {
			target = config.VisibleAreas[realIndex]
		} else {
			return fmt.Errorf("invalid area number")
		}
	}

	fmt.Printf("Exploring %s...\n", target)

	locationDetail, err := config.Pokeapi.GetLocationArea(target)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range locationDetail.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandMap(cfg *Config, args []string) error {
	resp, err := cfg.Pokeapi.GetLocationAreas(cfg.NextURL)
	if err != nil {
		return err
	}

	cfg.NextURL = resp.Next
	cfg.PreviousURL = resp.Previous

	cfg.VisibleAreas = []string{}
	for i, area := range resp.Results {
		cfg.VisibleAreas = append(cfg.VisibleAreas, area.Name)
		fmt.Printf("%d. %s\n", i+1, area.Name)
	}
	return nil
}

func commandMapb(config *Config, args []string) error {
	if config.PreviousURL == nil {
		fmt.Println("You're on the first page")
		return nil
	}

	locationResp, err := config.Pokeapi.GetLocationAreas(config.PreviousURL)
	if err != nil {
		return fmt.Errorf("failed to fetch location areas: %v", err)
	}

	config.NextURL = locationResp.Next
	config.PreviousURL = locationResp.Previous

	fmt.Println("Location Areas:")
	for _, area := range locationResp.Results {
		fmt.Println(area.Name)
	}

	return nil
}

func commandRight(cfg *Config, args []string) error {
	return commandMap(cfg, args)
}

func commandLeft(cfg *Config, args []string) error {
	return commandMapb(cfg, args)
}

func commandCatch(cfg *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: encounter <pokemon_name>")
	}

	// 1. Fetch wild pokemon data
	name := args[0]
	wildBase, err := cfg.Pokeapi.GetPokemon(name)
	if err != nil {
		return err
	}

	// 2. Create Wild Instance (Random level 1-10)
	wildMon, err := game.NewBattlePokemon(wildBase, rand.Intn(5)+1, cfg.Pokeapi)
	if err != nil {
		return fmt.Errorf("failed to create pokemon: %w", err)
	}

	fmt.Printf("A wild %s appeared!\n", wildBase.Name)

	// 3. Start Battle Loop
	// We pass the party, the wild mon, and the inventory
	// We also pass the PC so the battle engine can add the pokemon there if the party is full
	caught := game.StartBattle(cfg.Party, wildMon, &cfg.Inventory, cfg.Pokeapi)

	if caught {
		// If the boolean returned true, it means the catch was successful
		// We add the base data to our CaughtPokemon (the Pokedex tracker)
		cfg.CaughtPokemon[name] = wildBase

		// Logic check: StartBattle should have handled appending 'wildMon'
		// to either cfg.Party or cfg.PC already.
	}

	// 4. Save the entire game state (Party, PC, Inventory, Pokedex)
	err = saveGame(cfg)
	if err != nil {
		return fmt.Errorf("battle finished but failed to save: %w", err)
	}

	return nil
}

func commandInspect(cfg *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: inspect <pokemon_name>")
	}
	name := args[0]

	// 1. Check Party (Active Team) - SHOWS PROGRESS
	for _, p := range cfg.Party {
		if p.Base.Name == name || p.Nickname == name {
			printBattlePokemonDetails(p, "Party")
			return nil
		}
	}

	// 2. Check PC (Storage) - SHOWS PROGRESS
	// (Assumes you have a PC slice in your config)
	for _, p := range cfg.PC {
		if p.Base.Name == name || p.Nickname == name {
			printBattlePokemonDetails(p, "PC Storage")
			return nil
		}
	}

	// 3. Fallback: Pokedex Data (Generic Species Info)
	pokemon, ok := cfg.CaughtPokemon[name]
	if !ok {
		return fmt.Errorf("you have not caught that pokemon")
	}

	fmt.Printf("--- Pokedex Entry: %s ---\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	// Added Base Experience info here
	fmt.Printf("Base XP Yield: %d\n", pokemon.BaseExperience)

	fmt.Println("Base Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %v\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, typeInfo := range pokemon.Types {
		fmt.Printf("  - %s\n", typeInfo.Type.Name)
	}

	return nil
}

// Helper function to print the detailed "RPG Style" view
func printBattlePokemonDetails(p *game.BattlePokemon, location string) {
	fmt.Printf("--- Inspected: %s (%s) ---\n", p.Nickname, location)
	fmt.Printf("Lvl: %d\n", p.Level)
	fmt.Printf("HP:  %d/%d\n", p.Stats.HP, p.Stats.MaxHP)

	// HERE IS THE XP INDICATION YOU WANTED
	fmt.Printf("XP:  %d / %d\n", p.XP, p.NextLevelXP)

	fmt.Printf("Status: %s\n", p.Status)
	fmt.Printf("Nature: %s\n", p.Nature)

	fmt.Println("Stats:")
	fmt.Printf("  -Attack:  %d\n", p.Stats.Attack)
	fmt.Printf("  -Defense: %d\n", p.Stats.Defense)
	fmt.Printf("  -Speed:   %d\n", p.Stats.Speed)

	fmt.Println("Moves:")
	for _, m := range p.Moves {
		fmt.Printf("  - %s (%s) Pwr:%d PP:%d/%d\n", m.Name, m.Type, m.Power, m.CurrentPP, m.MaxPP)
	}
}

func commandTeam(cfg *Config, args []string) error {
	if len(cfg.Party) == 0 {
		fmt.Println("You have no Pokemon in your team.")
		return nil
	}
	fmt.Println("--- Active Team ---")
	for i, p := range cfg.Party {
		// Go now calls p.Status.String() automatically because of the code above
		fmt.Printf("%d. %-12s | Lvl %d | HP: %d/%d | Status: %s\n",
			i+1, p.Nickname, p.Level, p.Stats.HP, p.Stats.MaxHP, p.Status)

		// XP Bar visual (optional but cool)
		fmt.Printf("   XP: %d/%d\n", p.XP, p.NextLevelXP)
	}
	return nil
}

func commandAddTeam(config *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: addteam <pokemon_name>")
	}
	name := args[0]

	baseData, ok := config.CaughtPokemon[name]
	if !ok {
		return fmt.Errorf("you haven't caught %s yet", name)
	}

	if len(config.Party) >= 6 {
		return fmt.Errorf("your party is full")
	}

	// Fix: Use 'baseData', pass 'config.Pokeapi', and handle the error
	// We default to level 5 for team additions from the box, or you could track level in CaughtPokemon later
	newMember, err := game.NewBattlePokemon(baseData, 5, config.Pokeapi)
	if err != nil {
		return fmt.Errorf("failed to create team member: %w", err)
	}

	config.Party = append(config.Party, newMember)

	fmt.Printf("%s added to your party!\n", name)
	return saveGame(config) // Good practice to save after modifying team
}

func commandBag(cfg *Config, args []string) error {
	fmt.Println("--- Inventory ---")
	fmt.Printf("Pokeballs: %d\n", cfg.Inventory.Pokeballs)
	fmt.Printf("Potions:   %d\n", cfg.Inventory.Potions)
	return nil
}

func commandHeal(cfg *Config, args []string) error {
	if len(cfg.Party) == 0 {
		return fmt.Errorf("you have no Pokemon to heal")
	}

	fmt.Println("Welcome to the Pokemon Center!")
	fmt.Println("Healing your team... Please wait...")

	time.Sleep(1 * time.Second)

	for _, p := range cfg.Party {
		p.HealFull() // This uses the helper we added in internal/game/models.go
	}

	fmt.Println("Your Pokemon are fighting fit! We hope to see you again!")

	// Save the game so their HP stays full if they quit
	return saveGame(cfg)
}

func commandEvolve(cfg *Config, args []string) error {
	if len(cfg.Party) == 0 {
		return fmt.Errorf("you have no Pokemon in your party")
	}

	// 1. Select Pokemon
	fmt.Println("--- Evolution Chamber ---")
	for i, p := range cfg.Party {
		fmt.Printf("%d. %s (Lvl %d)\n", i+1, p.Nickname, p.Level)
	}
	fmt.Println("c. Cancel")

	reader := bufio.NewScanner(os.Stdin)
	fmt.Print("Select Pokemon to evolve > ")
	reader.Scan()
	choice := reader.Text()

	if choice == "c" {
		return nil
	}
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(cfg.Party) {
		return fmt.Errorf("invalid selection")
	}

	selectedMon := cfg.Party[idx-1]
	fmt.Printf("Checking genetics for %s...\n", selectedMon.Nickname)

	// 2. FETCH EVOLUTION DATA (Replaces the Registry lookup)
	// A. Get Species to find the Chain URL
	species, err := cfg.Pokeapi.GetPokemonSpecies(selectedMon.Base.Name)
	if err != nil {
		return fmt.Errorf("cannot find species data: %w", err)
	}

	// B. Get the actual Chain
	chainData, err := cfg.Pokeapi.GetEvolutionChain(species.EvolutionChain.URL)
	if err != nil {
		return fmt.Errorf("cannot find evolution chain: %w", err)
	}

	// C. Find the next stage using a helper function
	nextStageName, minLevel, itemReq := game.FindNextEvolution(chainData.Chain, selectedMon.Base.Name)

	if nextStageName == "" {
		fmt.Printf("%s cannot evolve any further.\n", selectedMon.Nickname)
		return nil
	}

	// 3. CHECK REQUIREMENTS
	// Check Level
	if minLevel > 0 && selectedMon.Level < minLevel {
		fmt.Printf("%s needs to be at least level %d to evolve into %s.\n", selectedMon.Nickname, minLevel, nextStageName)
		return nil
	}

	// Check Item (if required)
	if itemReq != "" {
		count := cfg.Inventory.EvolutionStones[itemReq]
		if count <= 0 {
			fmt.Printf("You need a %s to evolve into %s.\n", itemReq, nextStageName)
			return nil
		}

		fmt.Printf("Evolution requires 1x %s. ", itemReq)
	}

	// 4. CONFIRMATION
	fmt.Printf("Evolve %s into %s? (y/n): ", selectedMon.Nickname, nextStageName)
	reader.Scan()
	if reader.Text() != "y" {
		fmt.Println("Evolution cancelled.")
		return nil
	}

	// 5. EXECUTE EVOLUTION
	fmt.Println("...")
	fmt.Println("What? " + selectedMon.Nickname + " is evolving!")

	// Fetch new base stats
	newBase, err := cfg.Pokeapi.GetPokemon(nextStageName)
	if err != nil {
		return fmt.Errorf("failed to fetch new form data: %w", err)
	}

	// Consume Item if used
	if itemReq != "" {
		cfg.Inventory.EvolutionStones[itemReq]--
	}

	// Pass the API client so it can fetch a new move!
	selectedMon.Evolve(newBase, cfg.Pokeapi)

	fmt.Printf("Congratulations! Your Pokemon evolved into %s!\n", selectedMon.Nickname)
	saveGame(cfg)
	return nil
}

func commandShop(cfg *Config, args []string) error {
	shopItems := map[string]int{
		"pokeball":    10,
		"greatball":   100,
		"ultraball":   500,
		"potion":      300,
		"superpotion": 700,
		"revive":      500,
		"fire-stone":  700,
		"water-stone": 700,
		"leaf-stone":  700,
	}

	if len(args) == 0 {
		fmt.Printf("--- Welcome to the PokeMart! --- (Balance: ₽%d)\n", cfg.Inventory.Money)
		for item, price := range shopItems {
			fmt.Printf("- %-12s: ₽%d\n", item, price)
		}
		fmt.Println("\nUsage: shop buy <item_name>")
		return nil
	}

	if args[0] == "buy" && len(args) == 2 {
		itemName := strings.ToLower(args[1])
		price, exists := shopItems[itemName]
		if !exists {
			return fmt.Errorf("we don't sell %s here", itemName)
		}

		if cfg.Inventory.Money < price {
			return fmt.Errorf("you don't have enough money! (Needs ₽%d)", price)
		}

		// Deduct money and add item
		cfg.Inventory.Money -= price
		switch itemName {
		case "pokeball":
			cfg.Inventory.Pokeballs++
		case "greatball":
			cfg.Inventory.Greatballs++
		case "ultraball":
			cfg.Inventory.Ultraballs++
		case "potion":
			cfg.Inventory.Potions++
		case "superpotion":
			cfg.Inventory.SuperPotions++
		case "fire-stone", "water-stone", "leaf-stone":
			if cfg.Inventory.EvolutionStones == nil {
				cfg.Inventory.EvolutionStones = make(map[string]int)
			}
			cfg.Inventory.EvolutionStones[itemName]++
		}

		fmt.Printf("Purchased %s! New balance: ₽%d\n", itemName, cfg.Inventory.Money)
		saveGame(cfg)
		return nil
	}

	return fmt.Errorf("invalid shop command")
}

func commandBattle(config *Config, args []string) error {
	if len(config.Party) == 0 {
		return fmt.Errorf("you have no Pokemon in your party! Use 'addteam <name>' first")
	}
	if len(args) != 1 {
		return fmt.Errorf("usage: battle <pokemon_name>")
	}

	enemyName := args[0]
	fmt.Printf("A wild %s appeared!\n", enemyName)

	enemyBase, err := config.Pokeapi.GetPokemon(enemyName)
	if err != nil {
		return err
	}

	enemyLevel := rand.Intn(5) + 1

	// Fix: Pass config.Pokeapi and handle the error return
	wildPokemon, err := game.NewBattlePokemon(enemyBase, enemyLevel, config.Pokeapi)
	if err != nil {
		return fmt.Errorf("failed to prepare battle: %w", err)
	}

	// Start the battle
	game.StartBattle(config.Party, wildPokemon, &config.Inventory, config.Pokeapi)

	saveGame(config)
	return nil
}

func commandPokedex(config *Config, args []string) error {
	if len(config.CaughtPokemon) == 0 {
		fmt.Println("Your Pokedex is empty. Go catch some Pokemon!")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for name := range config.CaughtPokemon {
		fmt.Printf(" - %s\n", name)
	}

	return nil
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"explore": {
			name:        "explore <area_name>",
			description: "List all Pokemon in a given area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch <pokemon_name>",
			description: "Attempt to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect <pokemon_name>",
			description: "View details about a caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "See all the Pokemon you've caught",
			callback:    commandPokedex,
		},
		"map": {
			name:        "map",
			description: "Display next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display previous 20 location areas",
			callback:    commandMapb,
		},
		"left": {
			name:        "left",
			description: "Shortcut for mapb (previous areas)",
			callback:    commandLeft,
		},
		"right": {
			name:        "right",
			description: "Shortcut for map (next areas)",
			callback:    commandRight,
		},
		"team": {
			name:        "team",
			description: "Display your pokemon team",
			callback:    commandTeam,
		},
		"addteam": {
			name:        "addteam",
			description: "Add a pokemon to your team",
			callback:    commandAddTeam,
		},
		"bag": {
			name:        "bag",
			description: "Shows your items in your bag",
			callback:    commandBag,
		},
		"heal": {
			name:        "heal",
			description: "Heals your pokemon in a poke-center",
			callback:    commandHeal,
		},
		"evolve": {
			name:        "evolve",
			description: "Evolves your pokemon using stones",
			callback:    commandEvolve,
		},
		"shop": {
			name:        "shop",
			description: "Opens the shop",
			callback:    commandShop,
		},
		"battle": {
			name:        "battle",
			description: "Start a pokemon battle",
			callback:    commandBattle,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
}
