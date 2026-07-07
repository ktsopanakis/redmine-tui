# Redmine TUI — working notes for Claude

Terminal UI for managing Redmine issues. Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea)
(Elm architecture) + Lip Gloss. Module: `github.com/ktsopanakis/redmine-tui`.

## Environment & workflow (important)

- Repo lives at `~/Projects/ktsopanakis/redmine-tui`. Go is installed via `mise`
  (`go` on PATH, 1.26.x).
- **Build & install:** `go build -o ~/.local/bin/redmine-tui .` — the installed
  binary is on PATH. After any code change, rebuild it; the user must **restart
  the running TUI** to pick up the new binary.
- **Test:** `go test ./...`  ·  **Vet:** `go vet ./...`  ·  **Format:** `gofmt -w .`
  (always gofmt before committing — CI has a `gofmt -s -l` gate).
- **Run:** `redmine-tui` (full-screen TUI — run in a real terminal, not via a
  tool call). Flags: `--setup`, `--show-config`, `--alt-screen`.
- **Config:** `~/.config/redmine-tui/config.yaml` (mode 0600). Holds the Redmine
  URL, API key, and pane colors. The API key is a secret — **never commit it or
  paste it into repo files.** Redmine instance: `https://redmine.i3inc.ca`.

### Git flow the user prefers
Work on a `feature/<name>` branch off `main`, commit, push, then
`git checkout main && git merge --ff-only feature/<name> && git push origin main`,
and delete the branch locally + remote. Only commit/push **when the user asks**.
Commit trailer:
`Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`

## Architecture / layout

- `main.go` — flag parsing, config load, first-run setup, launches the Bubble Tea program.
- `config/` — `Settings` struct, YAML load/save, `PromptForRedmineSetup` (interactive).
- `api/redmine.go` — `Client` with `doRequest` (sets `X-Redmine-API-Key` header,
  30s timeout, errors on status ≥400). Types: `Issue`, `Journal`, `User`, `Status`,
  `Priority`, `Project`. Methods: `GetIssues`, `GetIssue` (incl. journals),
  `GetCurrentUser`, `GetProjects`, `GetUsers`, `UpdateIssue(id, map)`,
  `GetStatuses`, `GetPriorities`.
- `app/` — the Bubble Tea model (all UI logic):
  - `model.go` — the `Model` struct and the big `Update` key/message switch;
    `InitialModel`, `Init`.
  - `messages.go` — msg types + fetch commands (`fetchIssues`, `fetchIssueDetail`, …).
  - `edit.go` — `editableFields`, edit/mutation commands (`updateIssueMultiple`,
    `addNote`, `updateIssueStatus`, `updateIssueFields`), `userDisplayName`, edit footer.
  - `filters.go` — `getFilteredIssues`; `assigneeOption` + `quickFilteredAssignees`.
  - `panes.go` — `updatePaneContent` (left issue list + right detail render),
    ID→name helpers.
  - `list_render.go` — user/project selection list building + overlay.
  - `ui.go` — top-level `View`, footer menu items, and the overlay renderers
    (`renderNoteOverlay`, `renderDescEditor`, `renderStatusPicker`, `renderQuickActions`).
  - `help.go` — help-modal content.
- `ui/` — reusable, app-agnostic components: `header`, `footer`, `pane`, `list`,
  `modal` (`RenderModal`, `RenderInputModal`, `OverlayOnContent`, `OverlayInCorner`),
  `loading`, `styles`.

## Keymap (current)

Main view (not in an input mode):
`↑/↓` or `j/k` navigate · `Tab` switch panes · `Enter` load detail ·
`f` text filter · `m` My/All toggle · `r` reload · `u` filter by users ·
`p` filter by projects · `a` **quick-actions popup** · `e` edit fields ·
`s` **quick status picker** · `c` **add note** · `?` help · `q`/`Ctrl+C` quit.

- **`a` quick-actions popup** — status + assignee + note in one dialog, applied in
  one `UpdateIssue` PUT (only changed fields + non-empty note). Tab/Shift+Tab move
  between fields; Status cycles with ←/→; Assignee is type-to-filter (←/→ steps
  matches, includes "Unassigned"); Note is a textarea. Status/assignee preselect to
  current. Ctrl+S apply, Esc cancel.
- **`s` status picker** — list with current pre-highlighted; ↑/↓ or number `1-9`,
  Enter applies (single `status_id` PUT).
- **`c` note** — multi-line textarea; Ctrl+S posts (as a journal note), Esc cancels.
- **`e` edit mode** — fields: subject, description, status, priority, assigned_to,
  done_ratio, due_date. Tab/Enter next field; ↑/↓ cycle select fields; Ctrl+S saves
  all; Esc cancels. **Description is multi-line:** press Enter on it to open a
  textarea editor (Enter = newline, Ctrl+S apply, Esc cancel).

## Conventions & gotchas

- Elm architecture: `Model` is passed **by value**; `Update` returns
  `(tea.Model, tea.Cmd)`; the loop is single-threaded, so maps on the model are safe.
  Network I/O happens inside `tea.Cmd` closures that return messages.
- `issueUpdatedMsg` is the shared completion message for **all** mutations
  (edit save, note, status, quick-actions). Its handler clears edit/note/quick
  state and refreshes the issue list + current detail.
- **`pendingEdits` is keyed by field name, not by issue** — it MUST be cleared
  after a save (done in the `issueUpdatedMsg` handler) and only overlaid on the
  detail pane while `editMode` is true, or saved values "stick" onto other issues.
  Edit mode uses **dirty-tracking** (`editedFields`): only fields the user actually
  changed are committed — navigating past a field never rewrites it.
- Multi-line fields are edited only via the dedicated textarea, never the
  single-line `editInput` (which strips newlines / truncated at the old CharLimit).
- **Tests drive the real `Update` loop** with `tea.KeyMsg`/`tea.WindowSizeMsg` and
  assert on model state (see `app/model_test.go`). Send
  `tea.WindowSizeMsg{Width:120,Height:40}` first to initialise the panes — otherwise
  `updatePaneContent` hits `strings.Repeat(count<0)` and panics.

## Known issues / backlog (from a prior review — not yet fixed)

- **CI Go version mismatch:** `go.mod` says `go 1.25.4` but
  `.github/workflows/*.yml` pin `setup-go` to `1.23`. Should use `1.25.x` or
  `go-version-file: go.mod`.
- **Pagination:** all fetches use `limit=100, offset=0` and ignore `TotalCount`,
  so >100 issues/users/projects are silently dropped; multi-user/project filters
  run client-side over that capped set.
- **Swallowed errors:** only `issuesLoadedMsg` sets `m.err`; failures fetching
  current user / users / projects / statuses / priorities are ignored silently.
- **Width math** uses byte length (`len`) in places instead of display width, so
  non-ASCII (e.g. Greek) subjects can misalign borders/padding (cosmetic).
- **Keymap overlap** (intentional): `s`, `c`, and assignee-in-`e` overlap with the
  `a` popup — quick single actions vs. the combined dialog.

## Recent history (features added with the user)

`c` note · edit dirty-tracking fix · multi-line description editor + "sticky field"
& newline-display fixes · `s` status picker · `a` quick-actions popup.
(One-off: issue #3888's description had been truncated to 500 chars by the old
single-line editor; it was restored from the journal `old_value`.)
