package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/c-bata/go-prompt"
)

func completer(d prompt.Document) []prompt.Suggest {
	suggestions := []prompt.Suggest{
		{Text: "docker", Description: ""},
		{Text: "exit", Description: "Exit command prompt"},
	}

	word := d.GetWordBeforeCursorUntilSeparator(" ")
	return prompt.FilterHasPrefix(suggestions, word, true)
}

func main() {
run:
	dockerCommand := prompt.Input(">>> ",
		completer,
		prompt.OptionTitle("docker prompt"),
		prompt.OptionSelectedDescriptionTextColor(prompt.LightGray),
		prompt.OptionInputTextColor(prompt.Fuchsia),
		prompt.OptionPrefixBackgroundColor(prompt.Cyan))

	splittedDockerCommands := strings.Split(dockerCommand, " ")
	if splittedDockerCommands[0] == "exit" {
		os.Exit(0)
	}

	ps := exec.Command(splittedDockerCommands[0], splittedDockerCommands[1:]...)
	res, err := ps.Output()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(res))

	goto run
}
