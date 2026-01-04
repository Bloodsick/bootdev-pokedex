package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"

	pokeapi "github.com/Bloodisck/bootdev-pokedex/internal/pokeapi"
)

type Config struct {
	NextURL       *string
	PreviousURL   *string
	Pokeapi       pokeapi.Client
	CaughtPokemon map[string]pokeapi.Pokemon
	VisibleAreas  []string
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
}

func main() {
	pokeClient := pokeapi.NewClient(5*time.Second, 5*time.Minute)
	config := &Config{
		Pokeapi:       pokeClient,
		CaughtPokemon: loadPokedex(),
	}
	startRepl(config)
}

const saveFilePath = "pokedex_save.json"

func savePokedex(cfg *Config) error {
	data, err := json.Marshal(cfg.CaughtPokemon)
	if err != nil {
		return err
	}
	return os.WriteFile(saveFilePath, data, 0644)
}

func loadPokedex() map[string]pokeapi.Pokemon {
	data, err := os.ReadFile(saveFilePath)
	if err != nil {
		return make(map[string]pokeapi.Pokemon)
	}

	loaded := make(map[string]pokeapi.Pokemon)
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		return make(map[string]pokeapi.Pokemon)
	}
	return loaded
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
		"catch",
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

func commandCatch(config *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: catch <pokemon_name>")
	}

	name := args[0]
	pokemon, err := config.Pokeapi.GetPokemon(name)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", pokemon.Name)

	res := rand.Intn(pokemon.BaseExperience)
	if res > 40 {
		fmt.Printf("%s escaped!\n", pokemon.Name)
		return nil
	}

	fmt.Printf("%s was caught!\n", pokemon.Name)
	fmt.Println("You may now inspect it with the inspect command.")

	config.CaughtPokemon[name] = pokemon
	savePokedex(config)
	return nil
}

func commandInspect(config *Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: inspect <pokemon_name>")
	}

	name := args[0]

	pokemon, ok := config.CaughtPokemon[name]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, typeInfo := range pokemon.Types {
		fmt.Printf("  - %s\n", typeInfo.Type.Name)
	}

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
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
}
