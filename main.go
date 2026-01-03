package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	pokeapi "github.com/Bloodisck/bootdev-pokedex/internal/pokeapi"
)

type Config struct {
	NextURL     *string
	PreviousURL *string
	Pokeapi     pokeapi.Client
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config, []string) error
}

func main() {
	pokeClient := pokeapi.NewClient(5*time.Second, 5*time.Minute)

	config := &Config{
		Pokeapi: pokeClient,
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

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
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
