# Redmine TUI

A Terminal User Interface (TUI) for Redmine, built with Go and Bubble Tea.

## Features

- �️ **Split-screen interface** - Issues list on the left, details on the right
- 🔍 **Smart filtering** - Filter by "My Tasks", project, and more
- 📝 View comprehensive issue information
- ✏️ **Advanced inline editing** - Edit multiple fields without leaving the detail view
- 🔄 Change issue status with full status list selection
- ⚡ **Quick actions** - Close, set in progress, or reject tickets with one key
- 📊 Update completion percentage (done ratio)
- 💬 Add notes/comments when saving changes
- 📂 Switch between projects (including "All Projects" option)
- 🎯 Make multiple changes before saving
- ❓ Built-in help popup with all shortcuts
- ⌨️ Fully keyboard-driven interface with tab navigation

## Installation

### Prerequisites

- Go 1.21 or later
- A Redmine account with API access enabled

### Build from source

```bash
go build -o redmine-tui ./cmd/redmine-tui
```

Or install directly:

```bash
go install ./cmd/redmine-tui
```

## Configuration

Before using the application, you need to configure your Redmine connection:

```bash
./redmine-tui --setup
```

This will prompt you for:
- **Redmine URL**: Your Redmine instance URL (e.g., `https://redmine.example.com`)
- **API Key**: Your Redmine API key (found in your account settings)

The configuration is saved in `~/.redmine-tui.json`.

## Usage

Simply run the application:

```bash
./redmine-tui
```

### Keyboard Shortcuts

#### Split-Screen View (Normal Mode)
- `↑/↓` or `j/k`: Navigate through issues (auto-updates detail panel)
- `/`: Start filtering - type to search by issue subject (case-insensitive)
  - While filtering, all command keys are disabled to allow typing freely
  - Press `ESC` to clear filter, `Enter` to accept and navigate
- `e`: Enter edit mode for selected issue
- `p`: Switch projects
- `f`: Open filters popup (toggle "My Tasks" and "Open Issues")
- `?`: Show help with all shortcuts
- `r`: Refresh issue list
- **Quick Actions** (with confirmation):
  - `c`: Close ticket (asks for confirmation)
  - `i`: Set ticket to In Progress (asks for confirmation)
  - `x`: Reject ticket (asks for confirmation)
- `q`: Quit

#### Split-Screen View (Edit Mode)
- `tab`: Move to next editable field
- `shift+tab`: Move to previous editable field
- `↑/↓`: Navigate within status list (when status field is active)
- `enter`: Confirm status selection (when status field is active)
- `ctrl+s`: Save all pending changes with notes
- `esc`: Cancel editing and discard changes
- `q`: Quit

**Editable Fields (use tab to navigate):**
1. **Subject** - Issue title
2. **Description** - Detailed description
3. **Status** - Select from available statuses
4. **Done %** - Completion percentage (0-100)
5. **Notes** - Add comments about your changes

#### Project List View
- `↑/↓` or `j/k`: Navigate through projects
- `enter`: Select project (select "All Projects" to show all issues)
- `esc`: Go back to split-screen view
- `/`: Search/filter projects
- `q`: Quit

#### Filter Popup
- `1`: Toggle "My Tasks" filter (show only issues assigned to you)
- `2`: Toggle "Open Issues" filter (hide closed/resolved)
- `esc` or `f`: Close popup
- `q`: Quit

#### Help Popup
- `?`: Open help popup
- Any key: Close help popup

## Project Structure

```
redmine-tui/
├── cmd/
│   └── redmine-tui/
│       └── main.go          # Entry point
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── redmine/
│   │   └── client.go        # Redmine API client
│   └── ui/
│       └── model.go         # Bubble Tea UI model
├── go.mod
└── README.md
```

## API Coverage

Currently supported Redmine API endpoints:

- `GET /issues.json` - List issues
- `GET /issues/:id.json` - Get issue details
- `PUT /issues/:id.json` - Update issue
- `GET /projects.json` - List projects
- `GET /issue_statuses.json` - List issue statuses

## Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.

## License

MIT License
