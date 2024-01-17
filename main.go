package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"pokedexcli/internal/pokecache"
	"strings"
	"time"
)

var expRange = []int{20, 608}

type clicommand struct {
	name     string
	message  string
	callback func(string, *config, map[string]pokemondata) error
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

type pokemondata struct {
	Abilities []struct {
		Ability struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"ability"`
		IsHidden bool `json:"is_hidden"`
		Slot     int  `json:"slot"`
	} `json:"abilities"`
	BaseExperience int `json:"base_experience"`
	Forms          []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"forms"`
	GameIndices []struct {
		GameIndex int `json:"game_index"`
		Version   struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"version"`
	} `json:"game_indices"`
	Height    int `json:"height"`
	HeldItems []struct {
		Item struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"item"`
		VersionDetails []struct {
			Rarity  int `json:"rarity"`
			Version struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version"`
		} `json:"version_details"`
	} `json:"held_items"`
	ID                     int    `json:"id"`
	IsDefault              bool   `json:"is_default"`
	LocationAreaEncounters string `json:"location_area_encounters"`
	Moves                  []struct {
		Move struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"move"`
		VersionGroupDetails []struct {
			LevelLearnedAt  int `json:"level_learned_at"`
			MoveLearnMethod struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"move_learn_method"`
			VersionGroup struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"version_group"`
		} `json:"version_group_details"`
	} `json:"moves"`
	Name          string `json:"name"`
	Order         int    `json:"order"`
	PastAbilities []any  `json:"past_abilities"`
	PastTypes     []any  `json:"past_types"`
	Species       struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"species"`
	Sprites struct {
		BackDefault      string `json:"back_default"`
		BackFemale       string `json:"back_female"`
		BackShiny        string `json:"back_shiny"`
		BackShinyFemale  string `json:"back_shiny_female"`
		FrontDefault     string `json:"front_default"`
		FrontFemale      string `json:"front_female"`
		FrontShiny       string `json:"front_shiny"`
		FrontShinyFemale string `json:"front_shiny_female"`
		Other            struct {
			DreamWorld struct {
				FrontDefault string `json:"front_default"`
				FrontFemale  any    `json:"front_female"`
			} `json:"dream_world"`
			Home struct {
				FrontDefault     string `json:"front_default"`
				FrontFemale      string `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale string `json:"front_shiny_female"`
			} `json:"home"`
			OfficialArtwork struct {
				FrontDefault string `json:"front_default"`
				FrontShiny   string `json:"front_shiny"`
			} `json:"official-artwork"`
			Showdown struct {
				BackDefault      string `json:"back_default"`
				BackFemale       string `json:"back_female"`
				BackShiny        string `json:"back_shiny"`
				BackShinyFemale  any    `json:"back_shiny_female"`
				FrontDefault     string `json:"front_default"`
				FrontFemale      string `json:"front_female"`
				FrontShiny       string `json:"front_shiny"`
				FrontShinyFemale string `json:"front_shiny_female"`
			} `json:"showdown"`
		} `json:"other"`
		Versions struct {
			GenerationI struct {
				RedBlue struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"red-blue"`
				Yellow struct {
					BackDefault      string `json:"back_default"`
					BackGray         string `json:"back_gray"`
					BackTransparent  string `json:"back_transparent"`
					FrontDefault     string `json:"front_default"`
					FrontGray        string `json:"front_gray"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"yellow"`
			} `json:"generation-i"`
			GenerationIi struct {
				Crystal struct {
					BackDefault           string `json:"back_default"`
					BackShiny             string `json:"back_shiny"`
					BackShinyTransparent  string `json:"back_shiny_transparent"`
					BackTransparent       string `json:"back_transparent"`
					FrontDefault          string `json:"front_default"`
					FrontShiny            string `json:"front_shiny"`
					FrontShinyTransparent string `json:"front_shiny_transparent"`
					FrontTransparent      string `json:"front_transparent"`
				} `json:"crystal"`
				Gold struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"gold"`
				Silver struct {
					BackDefault      string `json:"back_default"`
					BackShiny        string `json:"back_shiny"`
					FrontDefault     string `json:"front_default"`
					FrontShiny       string `json:"front_shiny"`
					FrontTransparent string `json:"front_transparent"`
				} `json:"silver"`
			} `json:"generation-ii"`
			GenerationIii struct {
				Emerald struct {
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"emerald"`
				FireredLeafgreen struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"firered-leafgreen"`
				RubySapphire struct {
					BackDefault  string `json:"back_default"`
					BackShiny    string `json:"back_shiny"`
					FrontDefault string `json:"front_default"`
					FrontShiny   string `json:"front_shiny"`
				} `json:"ruby-sapphire"`
			} `json:"generation-iii"`
			GenerationIv struct {
				DiamondPearl struct {
					BackDefault      string `json:"back_default"`
					BackFemale       string `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  string `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      string `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale string `json:"front_shiny_female"`
				} `json:"diamond-pearl"`
				HeartgoldSoulsilver struct {
					BackDefault      string `json:"back_default"`
					BackFemale       string `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  string `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      string `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale string `json:"front_shiny_female"`
				} `json:"heartgold-soulsilver"`
				Platinum struct {
					BackDefault      string `json:"back_default"`
					BackFemale       string `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  string `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      string `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale string `json:"front_shiny_female"`
				} `json:"platinum"`
			} `json:"generation-iv"`
			GenerationV struct {
				BlackWhite struct {
					Animated struct {
						BackDefault      string `json:"back_default"`
						BackFemale       string `json:"back_female"`
						BackShiny        string `json:"back_shiny"`
						BackShinyFemale  string `json:"back_shiny_female"`
						FrontDefault     string `json:"front_default"`
						FrontFemale      string `json:"front_female"`
						FrontShiny       string `json:"front_shiny"`
						FrontShinyFemale string `json:"front_shiny_female"`
					} `json:"animated"`
					BackDefault      string `json:"back_default"`
					BackFemale       string `json:"back_female"`
					BackShiny        string `json:"back_shiny"`
					BackShinyFemale  string `json:"back_shiny_female"`
					FrontDefault     string `json:"front_default"`
					FrontFemale      string `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale string `json:"front_shiny_female"`
				} `json:"black-white"`
			} `json:"generation-v"`
			GenerationVi struct {
				OmegarubyAlphasapphire struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      string `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale string `json:"front_shiny_female"`
				} `json:"omegaruby-alphasapphire"`
				XY struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      string `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale string `json:"front_shiny_female"`
				} `json:"x-y"`
			} `json:"generation-vi"`
			GenerationVii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  any    `json:"front_female"`
				} `json:"icons"`
				UltraSunUltraMoon struct {
					FrontDefault     string `json:"front_default"`
					FrontFemale      string `json:"front_female"`
					FrontShiny       string `json:"front_shiny"`
					FrontShinyFemale string `json:"front_shiny_female"`
				} `json:"ultra-sun-ultra-moon"`
			} `json:"generation-vii"`
			GenerationViii struct {
				Icons struct {
					FrontDefault string `json:"front_default"`
					FrontFemale  string `json:"front_female"`
				} `json:"icons"`
			} `json:"generation-viii"`
		} `json:"versions"`
	} `json:"sprites"`
	Stats []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Slot int `json:"slot"`
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
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
		"catch": {
			name:     "catch",
			message:  "Attempt to catch a pokemon: catch <pokemon>",
			callback: commandCatch,
		},
		"inspect": {
			name:     "inspect",
			message:  "Inspect details of caught pokemon: inspect <pokemon>",
			callback: commandInspect,
		},
		"pokedex": {
			name:     "pokedex",
			message:  "List your caught pokemon",
			callback: commandPokedex,
		},
		"exit": {
			name:     "exit",
			message:  "Exit the pokedex",
			callback: commandExit,
		},
	}
}

func commandHelp(param string, conf *config, pokedex map[string]pokemondata) error {
	fmt.Println("\nWelcome to the Pokedex!\nUsage:")
	messages := "\n"
	for name, command := range commands {
		messages += fmt.Sprintf("%s: %s\n", name, command.message)
	}
	fmt.Println(messages)
	return nil
}

func commandPokedex(param string, conf *config, pokedex map[string]pokemondata) error {
	fmt.Println("Your pokedex:")
	for name := range pokedex {
		fmt.Printf(" - %s\n", name)
	}

	return nil
}

func commandInspect(param string, conf *config, pokedex map[string]pokemondata) error {
	pokemon, ok := pokedex[param]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	message := fmt.Sprintf("Name: %s\n", pokemon.Name)
	message += fmt.Sprintf("Height: %d\n", pokemon.Height)
	message += fmt.Sprintf("Weight: %d\n", pokemon.Weight)
	message += fmt.Sprintf("Stats:\n")
	for _, stat := range pokemon.Stats {
		message += fmt.Sprintf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	message += fmt.Sprintln("Types:")
	for _, pType := range pokemon.Types {
		message += fmt.Sprintf("  - %s\n", pType.Type.Name)
	}

	fmt.Println(message)

	return nil
}

func commandExplore(location string, conf *config, pokedex map[string]pokemondata) error {
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

func commandCatch(pokemonName string, conf *config, pokedex map[string]pokemondata) error {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", pokemonName)
	body, err := useCache(&url)
	if err != nil {
		return err
	}

	pokemon := pokemondata{}
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a pokeball at %s\n", pokemonName)
	catchChance := rand.Intn(200)
	if catchChance > pokemon.BaseExperience || catchChance > 195 {
		fmt.Printf("%s was caught!\n", pokemonName)
		pokedex[pokemonName] = pokemon
		fmt.Printf("You may now inspect %s\n", pokemonName)
	} else {
		fmt.Printf("%s escaped!\n", pokemonName)
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

func commandNextMap(param string, conf *config, pokedex map[string]pokemondata) error {
	return fetchLocations(conf.next, conf)
}

func commandPreviousMap(param string, conf *config, pokedex map[string]pokemondata) error {
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

func commandExit(param string, conf *config, pokedex map[string]pokemondata) error {
	return nil
}

var cache pokecache.Cache
var pokedex = map[string]pokemondata{}

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
			err := cmd.callback(param, &conf, pokedex)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nEncountered error completing command %s: %v\n\n", cmd.name, err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Encountered error reading input, quitting")
		}
	}
}
