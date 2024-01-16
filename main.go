package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

type clicommand struct {
	name     string
	message  string
	callback func(*config) error
}

type config struct {
	next     *string
	previous *string
}

type locationdata struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous any    `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
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
		"exit": {
			name:     "exit",
			message:  "Exit the pokedex",
			callback: commandExit,
		},
	}
}

func commandHelp(conf *config) error {
	fmt.Println("\nWelcome to the Pokedex!\nUsage:")
	messages := "\n"
	for name, command := range commands {
		messages += fmt.Sprintf("%s: %s\n", name, command.message)
	}
	fmt.Println(messages)
	return nil
}

func commandNextMap(conf *config) error {
	return fetchLocations(conf.next, conf)
}

func commandPreviousMap(conf *config) error {
	if conf.previous == nil {
		return errors.New("Already on the first page")
	}
	return fetchLocations(conf.previous, conf)
}

func fetchLocations(url *string, conf *config) error {
	res, err := http.Get(*url)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	locations := locationdata{}
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

	fmt.Println(next)
	fmt.Println(previous)
	*conf.next = next
	conf.previous = previous
	fmt.Println(*conf)

	return nil
}

func commandExit(conf *config) error {
	return nil
}

func main() {
	initCommands()
	next := "https://pokeapi.co/api/v2/location-area/"
	conf := config{
		next:     &next,
		previous: nil,
	}

	scanner := bufio.NewScanner(os.Stdin)
	running := true
	for running {
		fmt.Printf("pokedex> ")
		scanner.Scan()
		enteredCommand := scanner.Text()
		cmd, ok := commands[enteredCommand]
		if !ok {
			fmt.Println("Command not recognized")
		} else {
			if cmd.name == "exit" {
				running = false
			}
			err := cmd.callback(&conf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nEncountered error completing command %s: %v\n\n", cmd.name, err)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Encountered error reading input, quitting")
		}
	}
}
