package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	pokeapi "github.com/Bloodisck/bootdev-pokedex/internal/pokeapi"
)

type Config struct {
	NextURL       *string
	PreviousURL   *string
	Pokeapi       pokeapi.Client
	CaughtPokemon map[string]pokeapi.Pokemon
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
		CaughtPokemon: make(map[string]pokeapi.Pokemon),
	}
	startRepl(config)
}

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	pokemon := strings.Fields(text)
	return pokemon
}

func startRepl(config *Config) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Welcome to the Pokedex REPL!")

	for {
		fmt.Print("Pokedex > ")
		if !scanner.Scan() {
			if scanner.Err() != nil {
				fmt.Printf("Error reading input: %v\n", scanner.Err())
			} else {
				fmt.Println("\nGoodbye!")
			}
			break
		}
		userInput := scanner.Text()

		cleanedWords := cleanInput(userInput)

		if len(cleanedWords) == 0 {
			continue
		}

		commandName := cleanedWords[0]
		args := cleanedWords[1:]
		commands := getCommands()
		if command, exists := commands[commandName]; exists {
			err := command.callback(config, args)
			if err != nil {
				fmt.Printf("Error executing command '%s': %v\n", commandName, err)
			}
		} else {
			fmt.Printf("Unknown command: %s\n", commandName)
			fmt.Println("Type 'help' for available commands.")
		}

		fmt.Println()
	}

}

func commandExit(config *Config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(config *Config, args []string) error {
	commands := getCommands()

	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}

	return nil
}

func commandExplore(config *Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a location area name")
	}

	areaName := args[0]
	fmt.Printf("Exploring %s...\n", areaName)

	locationDetail, err := config.Pokeapi.GetLocationArea(areaName)
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range locationDetail.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandMap(config *Config, args []string) error {
	var urlToUse *string
	if config.NextURL != nil {
		urlToUse = config.NextURL
	}

	locationResp, err := config.Pokeapi.GetLocationAreas(urlToUse)
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
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}
}
