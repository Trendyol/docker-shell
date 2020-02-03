package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http"
	"net/url"

	commands "github.com/mstrYoda/docker-shell/lib"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/registry"
	"github.com/c-bata/go-prompt"
	"github.com/patrickmn/go-cache"
)

var dockerClient *docker.Client
var shellCommands commands.Commands = commands.New()

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

var commandExpression = regexp.MustCompile(`(?P<command>exec|stop|start|service|pull|run)\s{1}`)

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
		if len(completer.([]prompt.Suggest)) > 0 {
			memoryCache.Set(cacheKey, completer, cache.DefaultExpiration)
		}
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

		if command == "run" {
			if word == "-p" {
				return portMappingSuggestion()
			}

			return imagesSuggestion()
		}

		if command == "service" {
			return shellCommands.GetDockerServiceSuggestions()
		}

		if command == "pull" {

			if len(strings.Split(d.Text, " ")) > 2 {
				return []prompt.Suggest{}
			}

			if len(word) < 3 {
				return []prompt.Suggest{}
			}

			if word != command {
				return getFromCache(word)
			}
			return getFromCache("")
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

func portMappingSuggestion() []prompt.Suggest {
	images, _ := dockerClient.ImageList(context.Background(), types.ImageListOptions{All: true})
	suggestions := []prompt.Suggest{}

	for _, image := range images {
		inspection, _, _ := dockerClient.ImageInspectWithRaw(context.Background(), image.ID)

		exposedPortKeys := reflect.ValueOf(inspection.Config.ExposedPorts).MapKeys()

		for _, exposedPort := range exposedPortKeys {
			portAndType := strings.Split(exposedPort.String(), "/")
			port := portAndType[0]
			portType := portAndType[1]
			suggestions = append(suggestions, prompt.Suggest{Text: fmt.Sprintf("-p %s:%s/%s", port, port, portType), Description: inspection.RepoDigests[0]})
		}
	}

	return suggestions
}

func imagesSuggestion() []prompt.Suggest {
	images, _ := dockerClient.ImageList(context.Background(), types.ImageListOptions{All: true})
	suggestions := []prompt.Suggest{}

	for _, image := range images {
		ins, _, _ := dockerClient.ImageInspectWithRaw(context.Background(), image.ID)
		suggestions = append(suggestions, prompt.Suggest{Text: image.ID[7:], Description: ins.RepoDigests[0]})
	}

	return suggestions
}

func main() {
	dockerClient, _ = docker.NewEnvClient()
	if _, err := dockerClient.Ping(context.Background()); err != nil {
		fmt.Println("Couldn't check docker status please make sure docker is running.")
		fmt.Println(err)
		return
	}

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
