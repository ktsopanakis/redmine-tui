# Redmine TUI

Terminal UI for managing Redmine issues. Built with Go and Bubble Tea.

## Installation

Install using [mise](https://mise.jdx.dev/):

```bash
mise plugin add redmine-tui https://github.com/ktsopanakis/redmine-tui.git
mise install redmine-tui@0.0.9
mise use -g redmine-tui@0.0.9
```

Configure on first run:

```bash
redmine-tui --setup
```

## Development

Clone and build:

```bash
git clone https://github.com/ktsopanakis/redmine-tui.git
cd redmine-tui
go build
./redmine-tui --setup
```
