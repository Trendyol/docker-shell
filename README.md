# docker-shell

[![License: MIT](https://img.shields.io/badge/License-MIT-ligthgreen.svg)](https://opensource.org/licenses/MIT)

A simple interactive prompt for docker. Inspired from [kube-prompt](https://github.com/c-bata/kube-prompt) uses [go-prompt](https://github.com/c-bata/go-prompt).

[![asciicast](https://asciinema.org/a/3CjSyThXZMO5ocosaKkuJLlAF.svg)](https://asciinema.org/a/3CjSyThXZMO5ocosaKkuJLlAF)

<h4>Image suggestion from docker hub</h4>

[![asciicast](https://asciinema.org/a/UCfYZNXCcVxIiqNKsAMtEhmiM.svg)](https://asciinema.org/a/UCfYZNXCcVxIiqNKsAMtEhmiM)

<h4>Port mapping suggestion</h4>

[![asciicast](https://asciinema.org/a/7aWKWQJqqHZkpWZXwfy8AcrPj.svg)](https://asciinema.org/a/7aWKWQJqqHZkpWZXwfy8AcrPj)

Features:

* [X] Suggest docker commands
* [X] List container ids&names after docker exec/start/stop commands
* [ ] Suggest command parameters based on typed command
* [X] List images from docker hub after docker pull command [v1.2.0](https://github.com/Trendyol/docker-shell/milestone/1)
* [X] Suggest port mappings after docker run command [v1.3.0](https://github.com/Trendyol/docker-shell/milestone/2)
* [X] Suggest available images after docker run command [v1.3.0](https://github.com/Trendyol/docker-shell/milestone/2)


<h3>Installation</h3>

<b>Homebrew</b> :

  `brew tap trendyol/trendyol-tap`

  `brew install docker-shell`

After install it you can type `docker-shell` and run interactive shell.
