# docker-shell

[![License: MIT](https://img.shields.io/badge/License-MIT-ligthgreen.svg)](https://opensource.org/licenses/MIT)

A simple interactive prompt for docker. Inspired from [kube-prompt](https://github.com/c-bata/kube-prompt) uses [go-prompt](https://github.com/c-bata/go-prompt).

[![asciicast](https://asciinema.org/a/3CjSyThXZMO5ocosaKkuJLlAF.svg)](https://asciinema.org/a/3CjSyThXZMO5ocosaKkuJLlAF)

[![asciicast](https://asciinema.org/a/UCfYZNXCcVxIiqNKsAMtEhmiM.svg)](https://asciinema.org/a/UCfYZNXCcVxIiqNKsAMtEhmiM)

Features:

* [X] Suggest docker commands
* [X] List container ids&names after docker exec/start/stop commands
* [ ] Suggest command parameters based on typed command
* [X] List images from docker hub after docker pull command

<h3>Installation</h3>

<b>Homebrew</b> :

  `brew tap trendyol/trendyol-tap`

  `brew install docker-shell`

After install it you can type `docker-shell` and run interactive shell.
