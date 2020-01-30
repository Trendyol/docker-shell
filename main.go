package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"github.com/c-bata/go-prompt"
)

var dockerClient *docker.Client
var lastValidKeyword string

func isDockerCommand(kw string) bool {
	dockerCommands := []string{
		"docker",
		"attach",
		"build",
		"builder",
		"checkpoint",
		"commit",
		"config",
		"container",
		"context",
		"cp",
		"create",
		"diff",
		"events",
		"exec",
		"export",
		"history",
		"image",
		"images",
		"import",
		"info",
		"inspect",
		"kill",
		"load",
		"login",
		"logout",
		"logs",
		"manifest",
		"network",
		"node",
		"pause",
		"plugin",
		"port",
		"ps",
		"pull",
		"push",
		"rename",
		"restart",
		"rm",
		"rmi",
		"run",
		"save",
		"search",
		"secret",
		"service",
		"stack",
		"start",
		"stats",
		"stop",
		"swarm",
		"system",
		"tag",
		"top",
		"trust",
		"unpause",
		"update",
		"version",
		"volume",
		"wait",
	}

	for _, cmd := range dockerCommands {
		if cmd == kw {
			return true
		}
	}

	return false
}

func completer(d prompt.Document) []prompt.Suggest {
	word := d.GetWordBeforeCursor()

	for _, cmd := range strings.Split(d.Text, " ") {
		if strings.HasPrefix(cmd, "-") || strings.HasPrefix(cmd, "/") || cmd == "" {
			continue
		}

		lastValidKeyword = cmd
	}

	if word == "" {

		if lastValidKeyword == "exec" || lastValidKeyword == "stop" || lastValidKeyword == "port" {
			return containerListCompleter(false)
		}

		if lastValidKeyword == "start" {
			return containerListCompleter(true)
		}

		if lastValidKeyword == "service" {
			return dockerServiceCommandCompleter()
		}
	}

	suggestions := []prompt.Suggest{
		{Text: "attach", Description: "Attach local standard input, output, and error streams to a running container"},
		{Text: "build", Description: "Build an image from a Dockerfile"},
		{Text: "builder", Description: "Manage builds"},
		{Text: "checkpoint", Description: "Manage checkpoints"},
		{Text: "commit", Description: "Create a new image from a container’s changes"},
		{Text: "config", Description: "Manage Docker configs"},
		{Text: "container", Description: "Manage containers"},
		{Text: "context", Description: "Manage contexts"},
		{Text: "cp", Description: "Copy files/folders between a container and the local filesystem"},
		{Text: "create", Description: "Create a new container"},
		{Text: "diff", Description: "Inspect changes to files or directories on a container’s filesystem"},
		{Text: "events", Description: "Get real time events from the server"},
		{Text: "exec", Description: "Run a command in a running container"},
		{Text: "export", Description: "Export a container’s filesystem as a tar archive"},
		{Text: "history", Description: "Show the history of an image"},
		{Text: "image", Description: "Manage images"},
		{Text: "images", Description: "List images"},
		{Text: "import", Description: "Import the contents from a tarball to create a filesystem image"},
		{Text: "info", Description: "Display system-wide information"},
		{Text: "inspect", Description: "Return low-level information on Docker objects"},
		{Text: "kill", Description: "Kill one or more running containers"},
		{Text: "load", Description: "Load an image from a tar archive or STDIN"},
		{Text: "login", Description: "Log in to a Docker registry"},
		{Text: "logout", Description: "Log out from a Docker registry"},
		{Text: "logs", Description: "Fetch the logs of a container"},
		{Text: "manifest", Description: "Manage Docker image manifests and manifest lists"},
		{Text: "network", Description: "Manage networks"},
		{Text: "node", Description: "Manage Swarm nodes"},
		{Text: "pause", Description: "Pause all processes within one or more containers"},
		{Text: "plugin", Description: "Manage plugins"},
		{Text: "port", Description: "List port mappings or a specific mapping for the container"},
		{Text: "ps", Description: "List containers"},
		{Text: "pull", Description: "Pull an image or a repository from a registry"},
		{Text: "push", Description: "Push an image or a repository to a registry"},
		{Text: "rename", Description: "Rename a container"},
		{Text: "restart", Description: "Restart one or more containers"},
		{Text: "rm", Description: "Remove one or more containers"},
		{Text: "rmi", Description: "Remove one or more images"},
		{Text: "run", Description: "Run a command in a new container"},
		{Text: "save", Description: "Save one or more images to a tar archive (streamed to STDOUT by default)"},
		{Text: "search", Description: "Search the Docker Hub for images"},
		{Text: "secret", Description: "Manage Docker secrets"},
		{Text: "service", Description: "Manage services"},
		{Text: "stack", Description: "Manage Docker stacks"},
		{Text: "start", Description: "Start one or more stopped containers"},
		{Text: "stats", Description: "Display a live stream of container(s) resource usage statistics"},
		{Text: "stop", Description: "Stop one or more running containers"},
		{Text: "swarm", Description: "Manage Swarm"},
		{Text: "system", Description: "Manage Docker"},
		{Text: "tag", Description: "Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE"},
		{Text: "top", Description: "Display the running processes of a container"},
		{Text: "trust", Description: "Manage trust on Docker images"},
		{Text: "unpause", Description: "Unpause all processes within one or more containers"},
		{Text: "update", Description: "Update configuration of one or more containers"},
		{Text: "version", Description: "Show the Docker version information"},
		{Text: "volume", Description: "Manage volumes"},
		{Text: "wait", Description: "Block until one or more containers stop, then print their exit codes"},
		{Text: "exit", Description: "Exit command prompt"},
	}

	return prompt.FilterHasPrefix(suggestions, word, true)
}

func dockerServiceCommandCompleter() []prompt.Suggest {
	return []prompt.Suggest{
		{Text: "create", Description: "Create a new service"},
		{Text: "inspect", Description: "Display detailed information on one or more services"},
		{Text: "logs", Description: "Fetch the logs of a service or task"},
		{Text: "ls", Description: "List services"},
		{Text: "ps", Description: "List the tasks of one or more services"},
		{Text: "rm", Description: "Remove one or more services"},
		{Text: "rollback", Description: "Revert changes to a service’s configuration"},
		{Text: "scale", Description: "Scale one or multiple replicated services"},
		{Text: "update", Description: "Update a service"},
	}
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
run:
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

	goto run
}
