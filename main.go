package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"pokedexcli/internal/pokecache"
	"strings"
	"time"
)

type clicommand struct {
	name     string
	message  string
	callback func(string, *config) error
}

type config struct {
	next     *string
	previous *string
}

type locationlist struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type locationdata struct {
	EncounterMethodRates []struct {
		EncounterMethod struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"encounter_method"`
		VersionDetails []struct {
			Rate    int `json:"rate"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"encounter_method_rates"`
	GameIndex int `json:"game_index"`
	ID        int `json:"id"`
	Location  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"location"`
	Name  string `json:"name"`
	Names []struct {
		Language struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"language"`
		Name string `json:"name"`
	} `json:"names"`
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
		VersionDetails []struct {
			EncounterDetails []struct {
				Chance          int   `json:"chance"`
				ConditionValues []any `json:"condition_values"`
				MaxLevel        int   `json:"max_level"`
				Method          struct {
					Name string `json:"name"`
					URL  string `json:"url"`
				} `json:"method"`
				MinLevel int `json:"min_level"`
			} `json:"encounter_details"`
			MaxChance int `json:"max_chance"`
			Version   struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"pokemon_encounters"`
}

var commands = map[string]clicommand{}

func initCommands() {
	commands = map[string]clicommand{
		"help": {
			name:     "help",
			message:  "Displays a help message",
			callback: commandHelp,
		},
		"map": {
			name:     "map",
			message:  "Displays the next 20 map locations",
			callback: commandNextMap,
		},
		"mapb": {
			name:     "mapb",
			message:  "Displays the previous 20 map locations",
			callback: commandPreviousMap,
		},
		"explore": {
			name:     "explore",
			message:  "Choose a location to explore: explore <location>",
			callback: commandExplore,
		},
		"exit": {
			name:     "exit",
			message:  "Exit the pokedex",
			callback: commandExit,
		},
	}
}

func commandHelp(param string, conf *config) error {
	fmt.Println("\nWelcome to the Pokedex!\nUsage:")
	messages := "\n"
	for name, command := range commands {
		messages += fmt.Sprintf("%s: %s\n", name, command.message)
	}
	fmt.Println(messages)
	return nil
}

func commandExplore(location string, conf *config) error {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", location)
	body, err := useCache(&url)
	if err != nil {
		return err
	}

	locationData := locationdata{}
	err = json.Unmarshal(body, &locationData)
	if err != nil {
		return err
	}

	fmt.Printf("Exploring %s...\n", location)
	fmt.Println("Found Pokemon:")
	for _, pokemonEncounter := range locationData.PokemonEncounters {
		fmt.Printf(" - %s\n", pokemonEncounter.Pokemon.Name)
	}

	return nil
}

func useCache(url *string) ([]byte, error) {
	body, cached := cache.Get(*url)
	if !cached {
		res, err := http.Get(*url)
		if err != nil {
			return make([]byte, 0), err
		}

		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return make([]byte, 0), err
		}
		res.Body.Close()
		body = resBody

		cache.Add(*url, body)
	}

	return body, nil
}

func commandNextMap(param string, conf *config) error {
	return fetchLocations(conf.next, conf)
}

func commandPreviousMap(param string, conf *config) error {
	if conf.previous == nil {
		return errors.New("Already on the first page")
	}
	return fetchLocations(conf.previous, conf)
}

func fetchLocations(url *string, conf *config) error {
	body, err := useCache(url)
	locations := locationlist{}
	err = json.Unmarshal(body, &locations)
	if err != nil {
		return err
	}

	next := locations.Next
	var previous *string
	if prev, ok := locations.Previous.(string); ok {
		previous = &prev
	} else {
		previous = nil
	}
	for _, location := range locations.Results {
		fmt.Println(location.Name)
	}

	*conf.next = next
	conf.previous = previous
	fmt.Println(*conf)

	return nil
}

func commandExit(param string, conf *config) error {
	return nil
}

var cache pokecache.Cache

func main() {
	initCommands()
	next := "https://pokeapi.co/api/v2/location-area/"
	conf := config{
		next:     &next,
		previous: nil,
	}

	cache = pokecache.NewCache(30 * time.Second)

	scanner := bufio.NewScanner(os.Stdin)
	running := true
	for running {
		fmt.Printf("pokedex> ")
		scanner.Scan()

		enteredCommand := strings.Fields(scanner.Text())
		cmd, ok := commands[enteredCommand[0]]
		param := ""
		if len(enteredCommand) > 1 {
			param = enteredCommand[1]
		}

		if !ok {
			fmt.Println("Command not recognized")
		} else {
			if cmd.name == "exit" {
				running = false
			}
			err := cmd.callback(param, &conf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nEncountered error completing command %s: %v\n\n", cmd.name, err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Encountered error reading input, quitting")
		}
	}
}
