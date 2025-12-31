# Redmine TUI

A glamorous Terminal User Interface (TUI) application built with Go and Charm libraries for interacting with Redmine.

## Features

- **Dual Pane Layout**: 1/3 and 2/3 split with scrollable content
- **Adaptive Footer**: Dynamically hides items based on terminal width
- **Configurable Colors**: YAML-based configuration for styling
- **Redmine Integration**: Connect to your Redmine instance via API
- **Mouse Support**: Click to switch panes, scroll with mouse wheel, native text selection
- **Alt-Screen Mode**: Optional mode that clears output on exit
- **Help Modal**: Press `?` for comprehensive keyboard shortcuts
- **Edit Mode**: Modify issue fields directly from the TUI
- **Filtering**: Filter by users, projects, and text search

## Installation

### Using mise (Recommended)

Install [mise](https://mise.jdx.dev/) if you haven't already.

**Direct install from GitHub:**
```bash
mise install https://github.com/ktsopanakis/redmine-tui.git@v0.0.5
mise use -g redmine-tui@v0.0.5
```

Or install latest:
```bash
mise install https://github.com/ktsopanakis/redmine-tui.git@latest
```

**Clone and build:**
```bash
git clone https://github.com/ktsopanakis/redmine-tui.git
cd redmine-tui
mise trust               # Trust the mise config
mise install             # Installs Go 1.25.4
mise run install         # Builds and installs to $GOPATH/bin
```

**Direct install with Go:**
```bash
mise use -g go@1.25.4                                    # Install Go globally
go install github.com/ktsopanakis/redmine-tui@latest    # Install directly
```

### Manual Installation

Requires Go 1.25.4 or later:

```bash
git clone https://github.com/ktsopanakis/redmine-tui.git
cd redmine-tui
go build
```

## Development Tasks

With mise installed in the project directory:

```bash
mise run build        # Build binary
mise run install      # Install to $GOPATH/bin
mise run dev          # Build and run
mise run test         # Run tests
mise run clean        # Remove binary
```

## Quick Start

### First-Time Setup

Run the interactive setup to configure your Redmine connection:

```bash
./redmine-tui --setup
```

This will prompt you for:
- Redmine URL (e.g., `https://redmine.example.com`)
- API Key (found in your Redmine account settings)

The configuration will be saved to `~/.config/redmine-tui/config.yaml`

### Running

Once configured, simply run:

```bash
./redmine-tui
```

## Configuration

### Config File Location

The configuration file is stored at: `~/.config/redmine-tui/config.yaml`

To see the exact path on your system:

```bash
./redmine-tui --show-config
```

### Manual Configuration

You can also manually edit the config file:

```yaml
redmine:
  url: "https://redmine.example.com"
  api_key: "your_api_key_here"
colors:
  active_pane_border: "#FF00FF"
  inactive_pane_border: "#874BFD"
  header_background: "#7D56F4"
  header_text: "#FAFAFA"
  footer_background: "#3C3C3C"
  footer_text: "#FAFAFA"
```

## Key Bindings

- **Tab**: Switch between left and right panes
- **↑↓** or **jk**: Scroll up/down
- **PgUp/PgDn**: Page up/down
- **?**: Toggle help (in development)
- **q** or **Ctrl+C**: Quit
- **Mouse**: Click to switch panes, scroll with wheel, drag to select text

## Command-Line Options

```bash
# Run with default settings
./redmine-tui

# Run interactive setup
./redmine-tui --setup

# Show config file location
./redmine-tui --show-config

# Use alternate screen buffer (clears on exit)
./redmine-tui --alt-screen
```

## Project Structure

```
redmine-tui/
├── main.go       # Application entry point (orchestration)
├── config.go     # Settings struct and configuration loading
├── styles.go     # Lipgloss styles initialization
├── model.go      # Bubble Tea model, state, and update logic
├── ui.go         # View rendering and display logic
├── go.mod        # Go module dependencies
└── go.sum        # Dependency checksums
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) v1.3.10 - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) v1.1.0 - Styling library
- [Bubbles](https://github.com/charmbracelet/bubbles) v0.21.0 - UI components
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing

## Migration from Local config.yaml

If you have an old `config.yaml` in the project directory, the app now uses `~/.config/redmine-tui/config.yaml` instead. You can:

1. Run `./redmine-tui --setup` to create a new config with Redmine settings
2. Manually copy your color settings from the old `config.yaml` to the new location
3. The old `config.yaml` is now ignored by git and can be deleted

## Technical Details

### Adaptive Footer

The footer automatically adjusts displayed items based on terminal width:
- Always shows: "Tab: Switch" and "q: Quit" (or "v: Exit" and "y: Copy" in selection mode)
- Hides optional items when space is limited
- Uses plain text length measurement for accurate width calculation

### Text Selection

Text selection uses OSC 52 escape sequences for clipboard access:
- Works across SSH and local terminals
- Compatible with most modern terminal emulators
- Copies entire line at cursor position

### Code Organization

- **config.go**: Handles YAML configuration loading with sensible defaults
- **styles.go**: Initializes Lipgloss styles from loaded configuration
- **model.go**: Implements Bubble Tea's Model interface with state management
- **ui.go**: Handles all rendering logic including border title embedding
- **main.go**: Minimal orchestration - loads config, initializes styles, starts program

## Development

The codebase follows a clean architecture:
1. Configuration is loaded first
2. Styles are initialized from configuration
3. Model handles all state and update logic
4. UI handles all rendering and view logic
5. Main is just orchestration

This makes it easy to:
- Add new configuration options
- Customize styling
- Extend functionality
- Test components independently
