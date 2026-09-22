# cmdpp (Command Plus Plus)

A fast, keyboard-driven CLI and TUI tool to store, manage, and instantly trigger command-line shortcuts and workflows.

## Features

- **CLI and TUI Modes**: Use fast terminal subcommands or launch the interactive TUI picker.
- **Store & Manage**: Save long or complex commands with friendly alias names and optional descriptions.
- **Instant TUI Picker**: Launch `cmdpp` to search, preview, and hit `Enter` to run any command directly in your shell.
- **Built-in Search**: Real-time filtering across command names, scripts, and descriptions.
- **Clipboard Integration**: Press `c` in the TUI to copy any command to your system clipboard (Wayland / X11 / OSC 52 supported).
- **Pager Integration**: `cmdpp list` automatically pipes output through `more` when run in an interactive terminal.
- **Execution Tracking**: Tracks execution counts and last-run timestamps for each command.
- **In-TUI Add & Edit**: Add (`a`) or edit (`e`) commands right inside the TUI without leaving your keyboard.

---

## Installation & Location

The `cmdpp` binary is installed at `~/.local/bin/cmdpp` (which is in your `$PATH`).

Source directory: `~/Work/cmdpp`

To rebuild or reinstall at any time:
```bash
cd ~/Work/cmdpp
go build -o ~/.local/bin/cmdpp main.go
```

Configuration and stored commands are saved in JSON format at:
```
~/.config/cmdpp/commands.json
```

---

## Usage

### 1. Store a Command
```bash
cmdpp add <name> --cmd "<command>" [--desc "<description>"]
```
**Example:**
```bash
cmdpp add tunnel-one --cmd "ssh -N -L user@ssh.example.org"
cmdpp add tunnel-one --cmd "ssh -N -L user@ssh.example.org" --desc "SSH tunnel to jump host"
```

### 2. Run a Stored Command
```bash
cmdpp run <name> [extra args...]
```
**Example:**
```bash
cmdpp run tunnel-one
```
Any extra arguments are automatically passed along to the executed command.

### 3. Launch Interactive TUI Picker
Simply run `cmdpp` without arguments:
```bash
cmdpp
```
Select a command with `↑`/`↓` (or `j`/`k`) and press `Enter` to execute it directly.

#### TUI Keyboard Shortcuts
| Key | Action |
|---|---|
| `Enter` | Execute selected command |
| `/` | Focus search / filter bar |
| `↑` / `↓` or `k` / `j` | Move selection |
| `a` | Add a new command via dialog |
| `e` | Edit the selected command |
| `d` or `x` | Delete selected command (with confirmation) |
| `c` | Copy command to clipboard |
| `Esc` | Clear filter / cancel modal |
| `q` / `Ctrl+C` | Quit without running |

### 4. List Stored Commands
```bash
cmdpp list
```
Uses the `more` pager when running in an interactive terminal, or writes cleanly to stdout when piped.
To disable the pager explicitly:
```bash
cmdpp list --no-pager
```

### 5. Delete a Stored Command
```bash
cmdpp rm <name>
```
*(Aliases: `cmdpp remove <name>`, `cmdpp delete <name>`)*

### 6. Help & Documentation
```bash
cmdpp --help
```
