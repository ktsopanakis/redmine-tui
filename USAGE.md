# Redmine TUI - Usage Guide

## Quick Start

1. **Setup**: Run `./redmine-tui --setup` and enter your Redmine URL and API key
2. **Launch**: Run `./redmine-tui`
3. **Navigate**: Use arrow keys to browse your issues
4. **Edit**: Press `e` to edit the selected issue
5. **Save**: Make your changes, then press `Ctrl+S` to save

## Understanding the Interface

### Split-Screen Layout

```
┌─────────────────────┬────────────────────────────────┐
│                     │                                │
│   ISSUES LIST       │     ISSUE DETAILS              │
│   (40% width)       │     (60% width)                │
│                     │                                │
│   #123 - Task 1     │   #123 - Task 1                │
│   #124 - Task 2 ◄── │   Project: My Project          │
│   #125 - Task 3     │   Status: In Progress          │
│                     │   Description: ...             │
│                     │                                │
└─────────────────────┴────────────────────────────────┘
  Project: My Project | ↑/↓: navigate • e: edit • p: projects
```

- **Left Panel**: List of your assigned issues
- **Right Panel**: Detailed view of the selected issue
- **Bottom**: Context-sensitive help and current project

## Edit Mode Workflow

When you press `e` to edit an issue, the interface transforms into an interactive form:

### 1. Navigate Between Fields

- Press `Tab` to move to the next field
- Press `Shift+Tab` to move to the previous field
- The active field is highlighted in **orange** with "editing" indicator

### 2. Field-by-Field Editing

**Subject Field:**
```
Subject (editing):
[Type your new subject here_______________]
```
- Type directly to change the issue title

**Description Field:**
```
Description (editing):
[Type your description here_______________]
```
- Type or paste your new description

**Status Field:**
```
Status (use ↑/↓ to select, enter to confirm):
┌─────────────┐
│ New         │
│ In Progress │ ◄── Use arrows
│ Resolved    │
│ Closed      │
└─────────────┘
```
- Use `↑/↓` to navigate the status list
- Press `Enter` to select a status

**Done % Field:**
```
Done % (0-100):
[75___]
```
- Type a number between 0 and 100

**Notes Field:**
```
Notes (add comments about your changes):
[Updated the description and set to 75% complete___]
```
- Add comments that will be attached to your update

### 3. Review Pending Changes

Changes are marked with an asterisk `*` and shown in orange:

```
Status: In Progress *
Done: 75% *
Description: Updated description text *
```

The bottom bar shows: `EDIT MODE [3 pending changes]`

### 4. Save or Cancel

- `Ctrl+S`: Save all changes with your notes
- `Esc`: Cancel and discard all changes

## Common Workflows

### Workflow 1: Quick Status Update

1. Navigate to issue (↑/↓)
2. Press `e` to edit
3. Press `Tab` twice to reach Status field
4. Use ↑/↓ to select new status
5. Press `Enter` to confirm
6. Press `Tab` to Notes field
7. Type "Moving to next stage"
8. Press `Ctrl+S` to save

### Workflow 2: Update Progress

1. Select issue
2. Press `e`
3. Press `Tab` three times to Done % field
4. Type new percentage (e.g., "75")
5. Press `Tab` to Notes
6. Type update note
7. Press `Ctrl+S`

### Workflow 3: Comprehensive Update

1. Press `e` to edit
2. Update subject (if needed)
3. Press `Tab`, update description
4. Press `Tab`, change status
5. Press `Tab`, update done %
6. Press `Tab`, add detailed notes
7. Review all changes (marked with *)
8. Press `Ctrl+S` to save everything at once

### Workflow 4: Switch Projects

1. Press `p` to open project list
2. Use ↑/↓ or type `/` to search
3. Press `Enter` to select project
4. View issues for that project

## Tips and Tricks

### Efficient Navigation

- Use `j/k` (vim-style) or arrow keys
- Press `/` to filter/search in any list
- The detail panel updates automatically as you navigate

### Batch Editing

- You can make multiple changes before saving
- Each change is tracked and shown with `*`
- The counter shows how many changes are pending
- All changes are saved together when you press `Ctrl+S`

### Notes Best Practice

- Always add notes when making changes
- Notes help your team understand what changed and why
- Notes appear in the issue's activity/history

### Keyboard Efficiency

- Learn the tab navigation - it's much faster than using a mouse
- `Esc` always takes you back/cancels
- `q` always quits the application
- `r` refreshes the current issue list

## Troubleshooting

### Issue not updating?

- Make sure you pressed `Ctrl+S` (not just Enter)
- Check for error messages in red at the bottom
- Verify your API key has update permissions

### Can't see all content?

- Use ↑/↓ to scroll in the detail panel
- Resize your terminal for more space
- The left panel is 40% width, right is 60%

### Status list doesn't appear?

- Make sure you're in edit mode (press `e` first)
- Tab to the Status field
- If no statuses appear, check your Redmine API connection

### Changes not tracking?

- Only fields that actually change are tracked
- If you change a field back to its original value, it won't be tracked
- The asterisk `*` only appears on modified fields

## Advanced Usage

### API Key Permissions

Your API key needs these permissions:
- Read issues
- Update issues
- Read projects
- Read issue statuses

### Configuration File

Edit `~/.redmine-tui.json` directly if needed:
```json
{
  "redmine_url": "https://redmine.example.com",
  "api_key": "your-api-key-here"
}
```

### Multiple Configurations

To use different Redmine instances:
1. Create different config files
2. Use environment variable (future feature)
3. Or run setup again to switch

## Keyboard Reference Card

```
┌─────────────────────────────────────────────────┐
│ NAVIGATION                                      │
├─────────────────────────────────────────────────┤
│ ↑/↓, j/k        Navigate lists                  │
│ /               Search/Filter                   │
│ p               Switch projects                 │
│ r               Refresh                         │
│ q               Quit                            │
├─────────────────────────────────────────────────┤
│ EDITING                                         │
├─────────────────────────────────────────────────┤
│ e               Enter edit mode                 │
│ Tab             Next field                      │
│ Shift+Tab       Previous field                  │
│ Ctrl+S          Save changes                    │
│ Esc             Cancel/Go back                  │
├─────────────────────────────────────────────────┤
│ STATUS FIELD ONLY                               │
├─────────────────────────────────────────────────┤
│ ↑/↓             Navigate statuses               │
│ Enter           Confirm selection               │
└─────────────────────────────────────────────────┘
```
