package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/registry"

	"github.com/c-bata/go-prompt"
	"github.com/hashicorp/go-retryablehttp"
	commands "github.com/mstrYoda/docker-shell/lib"
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
	client := retryablehttp.NewClient()
	client.HTTPClient = &http.Client{
		Timeout: 1 * time.Second,
	}
	client.RetryWaitMin = client.HTTPClient.Timeout
	client.RetryWaitMax = client.HTTPClient.Timeout
	client.RetryMax = 3
	client.Logger = nil
	url := url.URL{
		Scheme:   "https",
		Host:     "registry.hub.docker.com",
		Path:     "/v2/repositories/library",
		RawQuery: "page=1&page_size=" + strconv.Itoa(count),
	}
	apiURL := url.String()
	response, err := client.Get(apiURL)
	if err != nil {
		return nil
	}

	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	searchResult := &DockerHubResult{}
	decoder.Decode(searchResult)
	if searchResult.Items == nil || len(searchResult.Items) <= 0 {
		return nil
	}

	return searchResult.Items
}

func imageFromContext(imageName string, count int) []registry.SearchResult {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ctxResponse, err := dockerClient.ImageSearch(ctx, imageName, types.ImageSearchOptions{Limit: count})
	if err != nil {
		return nil
	}

	if ctxResponse == nil || len(ctxResponse) <= 0 {
		return nil
	}

	return ctxResponse
}

func imageFetchCompleter(imageName string, count int) []prompt.Suggest {
	searchResult := []registry.SearchResult{}
	if imageName != "" {
		searchResult = imageFromContext(imageName, 10)
	} else {
		searchResult = imageFromHubAPI(10)
	}

	if searchResult == nil || len(searchResult) <= 0 {
		return nil
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

var commandExpression = regexp.MustCompile(`(?P<command>exec|stop|start|run|service create|service inspect|service logs|service ls|service ps|service rollback|service scale|service update|service|pull|attach|build|commit|cp|create|events|export|history|images|import|info|inspect|kill|load|login|logs|ps|push|restart|rm|rmi|save|search|stack|stats|update|version)\s{1}`)

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
		if completer.([]prompt.Suggest) == nil {
			return []prompt.Suggest{}
		}
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

		if command == "run" {
			if word == "-p" {
				return portMappingSuggestion()
			}

			return imagesSuggestion()
		}

		if command == "pull" {
			if strings.Index(word, ":") != -1 || strings.Index(word, "@") != -1 {
				return []prompt.Suggest{}
			}

			if word == "" || len(word) > 2 {
				if len(strings.Split(d.Text, " ")) > 2 {
					return []prompt.Suggest{}
				}
				return getFromCache(word)
			}

			return []prompt.Suggest{}
		}
		if val, ok := shellCommands.IsDockerSubCommand(command); ok {
			return prompt.FilterHasPrefix(val, word, true)
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
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := dockerClient.Ping(ctx); err != nil {
		fmt.Println("Couldn't check docker status please make sure docker is running.")
		fmt.Println(err)
		return
	}
	go getFromCache("")
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
