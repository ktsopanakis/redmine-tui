# Redmine TUI

A glamorous Terminal User Interface (TUI) application built with Go and Charm libraries.

## Features

- **Dual Pane Layout**: 1/3 and 2/3 split with scrollable content
- **Adaptive Footer**: Dynamically hides items based on terminal width
- **Text Selection**: Visual mode for selecting and copying text
- **Configurable Colors**: YAML-based configuration for styling
- **Mouse Support**: Click to switch panes, scroll with mouse wheel
- **Alt-Screen Mode**: Optional mode that clears output on exit

## Project Structure

```
redmine-tui/
├── main.go       # Application entry point (orchestration)
├── config.go     # Settings struct and configuration loading
├── styles.go     # Lipgloss styles initialization
├── model.go      # Bubble Tea model, state, and update logic
├── ui.go         # View rendering and display logic
├── config.yaml   # Color configuration file
├── go.mod        # Go module dependencies
└── go.sum        # Dependency checksums
```

## Key Bindings

### Normal Mode
- **Tab**: Switch between left and right panes
- **↑↓** or **jk**: Scroll up/down
- **PgUp/PgDn**: Page up/down
- **v**: Enter selection mode
- **?**: Toggle help (in development)
- **q** or **Ctrl+C**: Quit

### Selection Mode
- **↑↓←→** or **hjkl**: Move selection cursor
- **y**: Copy selected line to clipboard (using OSC 52)
- **v**: Exit selection mode

## Configuration

Edit `config.yaml` to customize colors:

```yaml
colors:
  active_pane_border: "#FF00FF"
  inactive_pane_border: "#874BFD"
  header_background: "#7D56F4"
  header_text: "#FAFAFA"
  footer_background: "#3C3C3C"
  footer_text: "#FAFAFA"
```

## Building

```bash
go build -o redmine-tui
```

## Running

```bash
# Default mode (output persists)
./redmine-tui

# Alt-screen mode (clears on exit)
./redmine-tui --alt-screen
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) v1.3.10 - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) v1.1.0 - Styling library
- [Bubbles](https://github.com/charmbracelet/bubbles) v0.21.0 - UI components
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing

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
