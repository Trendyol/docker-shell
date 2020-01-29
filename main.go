package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mstrYoda/docker-shell/lib"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"github.com/c-bata/go-prompt"
)

var dockerClient *docker.Client
var lastValidKeyword string
var shellCommands commands.Commands = commands.New()

func completer(d prompt.Document) []prompt.Suggest {
	word := d.GetWordBeforeCursor()

	for _, cmd := range strings.Split(d.Text, " ") {
		if strings.HasPrefix(cmd, "-") || strings.HasPrefix(cmd, "/") || cmd == "" {
			continue
		}

		lastValidKeyword = cmd
	}

	if word == "" {

		if lastValidKeyword == "exec" || lastValidKeyword == "stop" {
			return containerListCompleter(false)
		}

		if lastValidKeyword == "start" {
			return containerListCompleter(true)
		}

		if lastValidKeyword == "service" {
			return shellCommands.GetDockerServiceSuggestions()
		}
	}

	return prompt.FilterHasPrefix(shellCommands.GetDockerSuggestions(), word, true)
}

func containerListCompleter(all bool) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	ctx := context.Background()
	cList, _ := dockerClient.ContainerList(ctx, types.ContainerListOptions{All: all})

	for _, container := range cList {
		suggestions = append(suggestions, prompt.Suggest{Text: container.ID, Description: container.Image})
	}

	return suggestions
}

func main() {
	dockerClient, _ = docker.NewEnvClient()

	for {
		dockerCommand := prompt.Input(">>> docker ",
			completer,
			prompt.OptionTitle("docker prompt"),
			prompt.OptionSelectedDescriptionTextColor(prompt.Turquoise),
			prompt.OptionInputTextColor(prompt.Fuchsia),
			prompt.OptionPrefixBackgroundColor(prompt.Cyan))

		splittedDockerCommands := strings.Split(dockerCommand, " ")
		if splittedDockerCommands[0] == "exit" {
			os.Exit(0)
		}

		var ps *exec.Cmd

		if splittedDockerCommands[0] == "clear" {
			ps = exec.Command("clear")
		} else {
			ps = exec.Command("docker", splittedDockerCommands...)
		}

		res, err := ps.Output()

		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(string(res))
	}
}
