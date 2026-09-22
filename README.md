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

## Installation

Ensure `~/.local/bin` is in your `$PATH` (standard on modern Linux and macOS systems).

### Option 1: Automated Install (Linux & macOS - Recommended)

The installer automatically detects your operating system and architecture (`x86_64` or `arm64`), downloads the latest release with live progress, and installs the binary to `~/.local/bin/cmdpp`:

Using `curl`:
```bash
curl -fsSL https://raw.githubusercontent.com/hyuricane/cmdpp/master/install.sh | bash
```

Using `wget`:
```bash
wget -qO- https://raw.githubusercontent.com/hyuricane/cmdpp/master/install.sh | bash
```

*(Optional: You can specify a custom install directory by setting `BINDIR`, e.g.: `curl -fsSL ... | BINDIR=/usr/local/bin bash`)*

---

### Option 2: Manual Binary Download

If you prefer to download and extract pre-compiled binaries manually without running the script:

* **Linux (x86_64 / amd64):**
  ```bash
  mkdir -p ~/.local/bin
  curl -sSL https://github.com/hyuricane/cmdpp/releases/download/v0.1.0/cmdpp_0.1.0_linux_amd64.tar.gz | tar -xz -C ~/.local/bin ./cmdpp
  ```

* **Linux (ARM64 / aarch64):**
  ```bash
  mkdir -p ~/.local/bin
  curl -sSL https://github.com/hyuricane/cmdpp/releases/download/v0.1.0/cmdpp_0.1.0_linux_arm64.tar.gz | tar -xz -C ~/.local/bin ./cmdpp
  ```

* **macOS (Apple Silicon - M1/M2/M3/M4 / arm64):**
  ```bash
  mkdir -p ~/.local/bin
  curl -sSL https://github.com/hyuricane/cmdpp/releases/download/v0.1.0/cmdpp_0.1.0_darwin_arm64.tar.gz | tar -xz -C ~/.local/bin ./cmdpp
  ```

* **macOS (Intel / amd64):**
  ```bash
  mkdir -p ~/.local/bin
  curl -sSL https://github.com/hyuricane/cmdpp/releases/download/v0.1.0/cmdpp_0.1.0_darwin_amd64.tar.gz | tar -xz -C ~/.local/bin ./cmdpp
  ```

* **Windows & Checksums:**
  Download `.zip` packages for Windows (`amd64` / `arm64`) and `checksums.txt` from the [GitHub Releases Page](https://github.com/hyuricane/cmdpp/releases/latest).

---

### Option 2: Using `go install`
If you have Go (1.22+) installed:
```bash
go install github.com/hyuricane/cmdpp@latest
```

---

### Option 3: Build from Source
```bash
git clone https://github.com/hyuricane/cmdpp.git
cd cmdpp
make install
```
*(By default, installs to `~/.local/bin/cmdpp`. You can customize the location with `PREFIX=/usr/local make install`)*

---

## Configuration

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
