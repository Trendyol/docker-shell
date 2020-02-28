# docker-shell

A simple interactive prompt for Docker. Inspired from [kube-prompt](https://github.com/c-bata/kube-prompt) uses [go-prompt](https://github.com/c-bata/go-prompt).

[![License: MIT](https://img.shields.io/badge/License-MIT-ligthgreen.svg)](https://opensource.org/licenses/MIT) [![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v1.4%20adopted-ff69b4.svg)](CONTRIBUTING.md)

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
## Table Of Contents

- [Features:](#features)
- [Installation](#installation)
  - [Homebrew](#homebrew)
  - [Build From Source Code](#build-from-source-code)
- [How To Use](#how-to-use)
- [How To Contribute](#how-to-contribute)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Features

- [X] Suggest docker commands
- [X] List container ids&names after docker exec/start/stop commands
- [ ] Suggest command parameters based on typed command
- [X] List images from docker hub after docker pull command [v1.2.0](https://github.com/Trendyol/docker-shell/milestone/1)
- [X] Suggest port mappings after docker run command [v1.3.0](https://github.com/Trendyol/docker-shell/milestone/2)
- [X] Suggest available images after docker run command [v1.3.0](https://github.com/Trendyol/docker-shell/milestone/2)

## Installation

### Homebrew

You can install by using *homebrew*:

```bash
brew tap trendyol/trendyol-tap

brew install docker-shell
```

### Build From Source Code

You can build the command from source code by following the steps below:

```bash
git clone git@github.com:Trendyol/docker-shell.git

cd docker-shell

sudo go build -o /usr/local/bin/docker-shell .

docker-shell
```

## How To Use

After installation, you can type `docker-shell` and run the interactive shell.

[![asciicast](https://asciinema.org/a/AKDTBnD3gKKzACDdj7Tm670PJ.svg)](https://asciinema.org/a/AKDTBnD3gKKzACDdj7Tm670PJ)

Image suggestion from docker hub:

[![asciicast](https://asciinema.org/a/UCfYZNXCcVxIiqNKsAMtEhmiM.svg)](https://asciinema.org/a/UCfYZNXCcVxIiqNKsAMtEhmiM)

Port mapping suggestion:

[![asciicast](https://asciinema.org/a/7aWKWQJqqHZkpWZXwfy8AcrPj.svg)](https://asciinema.org/a/7aWKWQJqqHZkpWZXwfy8AcrPj)

## How To Contribute

Contributions are **welcome** and will be fully **credited**.

Please read the [CONTRIBUTING](CONTRIBUTING.md) and [CODE_OF_CONDUCT](CODE_OF_CONDUCT) files for details.
