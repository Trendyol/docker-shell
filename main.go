package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http"
	"net/url"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/registry"
	"github.com/c-bata/go-prompt"
	"github.com/patrickmn/go-cache"
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

//DockerHubResult : Wrap DockerHub API call
type DockerHubResult struct {
	PageCount        *int                    `json:"num_pages,omitempty"`
	ResultCount      *int                    `json:"num_results,omitempty"`
	ItemCountPerPage *int                    `json:"page_size,omitempty"`
	CurrentPage      *int                    `json:"page,omitempty"`
	Query            *string                 `json:"query,omitempty"`
	Items            []registry.SearchResult `json:"results,omitempty"`
}

func imageFromHubAPI(count int) []registry.SearchResult {
	url := url.URL{
		Scheme:   "https",
		Host:     "registry.hub.docker.com",
		Path:     "/v2/repositories/library",
		RawQuery: "page=1&page_size=" + strconv.Itoa(count),
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	apiURL := url.String()

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return nil
	}

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	searchResult := &DockerHubResult{}
	decoder.Decode(searchResult)
	return searchResult.Items
}

func imageFromContext(imageName string, count int) []registry.SearchResult {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctxResponse, err := dockerClient.ImageSearch(ctx, imageName, types.ImageSearchOptions{Limit: count})
	if err != nil {
		return nil
	}
	return ctxResponse
}

func imageFetchCompleter(imageName string, count int) []prompt.Suggest {
	searchResult := []registry.SearchResult{}
	if imageName == "" {
		searchResult = imageFromHubAPI(10)
	} else {
		searchResult = imageFromContext(imageName, 10)
	}

	suggestions := []prompt.Suggest{}
	for _, s := range searchResult {
		description := "Not Official"
		if s.IsOfficial {
			description = "Official"
		}
		suggestions = append(suggestions, prompt.Suggest{Text: s.Name, Description: "(" + description + ") " + s.Description})
	}
	return suggestions
}

var commandExpression = regexp.MustCompile(`(?P<command>exec|stop|start|service|pull)\s{1}`)

func getRegexGroups(text string) map[string]string {
	if !commandExpression.Match([]byte(text)) {
		return nil
	}

	match := commandExpression.FindStringSubmatch(text)
	result := make(map[string]string)
	for i, name := range commandExpression.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

var memoryCache = cache.New(5*time.Minute, 10*time.Minute)

func getFromCache(word string) []prompt.Suggest {
	cacheKey := "all"
	if word != "" {
		cacheKey = fmt.Sprintf("completer:%s", word)
	}
	completer, found := memoryCache.Get(cacheKey)
	if !found {
		completer = imageFetchCompleter(word, 10)
		memoryCache.Set(cacheKey, completer, cache.DefaultExpiration)
	}
	return completer.([]prompt.Suggest)
}

func completer(d prompt.Document) []prompt.Suggest {
	word := d.GetWordBeforeCursor()

	group := getRegexGroups(d.Text)
	if group != nil {
		command := group["command"]

		if command == "exec" || command == "stop" || command == "port" {
			return containerListCompleter(false)
		}

		if command == "start" {
			return containerListCompleter(true)
		}

		if command == "service" {
			return dockerServiceCommandCompleter()
		}

		if command == "pull" {
			if word != command {
				return getFromCache(word)
			}
			return getFromCache("")
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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := dockerClient.Info(ctx)
	if err != nil {
		fmt.Println("Couldn't check docker status please make sure docker is running.")
		fmt.Println(err)
		return
	}

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
