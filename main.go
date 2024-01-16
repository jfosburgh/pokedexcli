package main

import (
	"bufio"
	"fmt"
	"os"
)

type clicommand struct {
	name     string
	message  string
	callback func() error
}

var commands = map[string]clicommand{}

func initCommands() {
	commands = map[string]clicommand{
		"help": {
			name:     "help",
			message:  "Displays a help message",
			callback: commandHelp,
		},
		"exit": {
			name:     "exit",
			message:  "Exit the pokedex",
			callback: commandExit,
		},
	}
}

func commandHelp() error {
	fmt.Println("\nWelcome to the Pokedex!\nUsage:")
	messages := "\n"
	for name, command := range commands {
		messages += fmt.Sprintf("%s: %s\n", name, command.message)
	}
	fmt.Println(messages)
	return nil
}

func commandExit() error {
	return nil
}

func main() {
	initCommands()

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
			err := cmd.callback()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nEncountered error completing command %s\n\n", cmd.name)
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "Encountered error reading input, quitting")
		}
	}
}
