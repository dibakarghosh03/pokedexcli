package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dibakarghosh03/pokedexcli/internal/pokecache"
	"github.com/peterh/liner"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

type config struct {
	nextURL     *string
	previousURL *string
}

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Stat     struct {
			Name string `json:"name"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

type locationAreasResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type exploreResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

var cache *pokecache.Cache
var commands = map[string]cliCommand{}
var pokedex = map[string]Pokemon{}

func main() {
	cfg := &config{}
	cache = pokecache.NewCache(5 * time.Second)

	// Register Commands
	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Closes the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Displays the next 20 location areas",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Displays the previous 20 location areas",
			callback:    commandMapBack,
		},
		"explore": {
			name:        "exlore",
			description: "Displays the pokemon that can be found in the location area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempts to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Displays details about a pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Displays your pokedex",
			callback:    commandPokedex,
		},
	}

	// Create a liner instance
	line := liner.NewLiner()
	defer line.Close()

	// Configure liner
	line.SetCtrlCAborts(true)

	for {
		text, err := line.Prompt("Pokedex > ")
		if err != nil {
			fmt.Println("\nGoodbye!")
            return
		}

		// Clean it
		words := cleanInput(text)

		// If user typed nothing skip
		if len(words) == 0 {
			continue
		}

		cmdName := words[0]
		args := words[1:]

		// Lookup command in commands list
		cmd, exists := commands[cmdName]
		if !exists {
			fmt.Println("Unknown command:", cmdName)
			continue
		}

		// save to history
		line.AppendHistory(text)

		// Execute command
		if err := cmd.callback(cfg, args); err != nil {
			fmt.Println("Error: ", err)
		}
	}
}

func cleanInput(text string) []string {
	trimmed := strings.TrimSpace(text)
	lowered := strings.ToLower(trimmed)
	words := strings.Fields(lowered)

	return words
}

func commandExit(cfg *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0) // exits immediately
	return nil // never reached
}

func commandHelp(cfg *config, args []string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()

	for _, cmd := range commands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)
	}
	return nil
}

func commandMap(cfg *config, args []string) error {
	url := "https://pokeapi.co/api/v2/location-area?offset=0&limit=20"
	if cfg.nextURL != nil {
		url = *cfg.nextURL
	}

	data, err := fetchLocationAreas(url)
	if err != nil {
		return err
	}

	for _, loc := range data.Results {
		fmt.Println(loc.Name)
	}

	cfg.nextURL = data.Next
	cfg.previousURL = data.Previous

	return nil
}

func commandMapBack(cfg *config, args []string) error {
	if cfg.previousURL == nil {
		fmt.Println("You're on the first page!")
		return nil
	}

	data, err := fetchLocationAreas(*cfg.previousURL)
	if err != nil {
		return err
	}

	for _, loc := range data.Results {
		fmt.Println(loc.Name)
	}

	cfg.nextURL = data.Next
	cfg.previousURL = data.Previous

	return nil
}

func commandExplore(cfg *config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide a location area name")
	}

	areaName := args[0]
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s", areaName)

	fmt.Printf("Exploring %s...\n", areaName)

	// Try Cache
	var data exploreResponse
	if cached, ok := cache.Get(url); ok {
		if err := json.Unmarshal(cached, &data); err != nil {
			return err
		}
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		cache.Add(url, body)

		if err := json.Unmarshal(body, &data); err != nil {
			return err
		}
	}

	fmt.Println("Found Pokemon: ")
	for _, encounter := range data.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(cfg *config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: catch <pokemon_name>")
	}

	name := strings.ToLower(args[0])
	fmt.Printf("Throwing a Pokeball at %s...\n", name)

	// Check if already caught
	if _, exists := pokedex[name]; exists {
		fmt.Printf("You already caught %s!\n", name)
		return nil
	}

	// Fetch Pokemon data
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s", name)
	var pokemon Pokemon

	// Try cache first
	if cached, ok := cache.Get(url); ok {
		if err := json.Unmarshal(cached, &pokemon); err != nil {
			return err
		}
	} else {
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("bad status: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		cache.Add(url, body)

		if err := json.Unmarshal(body, &pokemon); err != nil {
			return err
		}
	}

	// Determine catch chance
	// Lower BaseExperience is easier to catch
	catchChance := 100 - pokemon.BaseExperience/3
	fmt.Printf("chance: %d%%\n", catchChance)
	if catchChance < 10 {
		catchChance = 10
	}

	if rand.Intn(100) < catchChance {
		fmt.Printf("%s was caught!\n", pokemon.Name)
		pokedex[name] = pokemon
	} else {
		fmt.Printf("%s escaped!\n", pokemon.Name)
	}

	return nil
}

func commandInspect(cfg *config, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: inspect <pokemon_name>")
	}

	name := strings.ToLower(args[0])

	// Check if player have caught pokemon
	pokemon, exists := pokedex[name]
	if !exists {
		return fmt.Errorf("you haven't caught %s yet", name)
	}

	// Print details
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)

	fmt.Println("Stats:")
	for _, s := range pokemon.Stats {
		fmt.Printf(" - %s: %d\n", s.Stat.Name, s.BaseStat)
	}

	fmt.Println("Types:")
	for _, t := range pokemon.Types {
		fmt.Printf("  - %s\n", t.Type.Name)
	}

	return nil
}

func commandPokedex(cfg *config, args []string) error {
	if pokedex == nil {
		fmt.Println("Your Pokedex is empty.")
		return nil
	}

	fmt.Println("Your Pokedex:")
	for _, pokemon := range pokedex {
		fmt.Printf(" - %s\n", pokemon.Name)
	}

	return nil
}

func fetchLocationAreas(url string) (*locationAreasResponse, error) {
	// 1. Try Cache
	if data, ok := cache.Get(url); ok {
		var result locationAreasResponse
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, err
		}
		return &result, nil
	}

	// 2. Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 3. Store in Cache
	cache.Add(url, body)

	var data locationAreasResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
