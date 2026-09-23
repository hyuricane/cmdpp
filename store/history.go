package store

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// DefaultHistoryFiles returns candidate shell history paths for the current user and OS,
// prioritized so that the current user's active shell ($SHELL) appears first.
func DefaultHistoryFiles() []string {
	var files []string
	seen := make(map[string]bool)

	addFile := func(path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		files = append(files, path)
	}

	// 1. Explicit $HISTFILE always takes highest precedence
	if hf := os.Getenv("HISTFILE"); hf != "" {
		addFile(hf)
	}

	home, _ := os.UserHomeDir()
	shell := strings.ToLower(filepath.Base(os.Getenv("SHELL")))

	var bashFiles, zshFiles, fishFiles, psFiles []string

	if home != "" {
		bashFiles = []string{
			filepath.Join(home, ".bash_history"),
		}
		zshFiles = []string{
			filepath.Join(home, ".zsh_history"),
			filepath.Join(home, ".histfile"),
		}
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			fishFiles = append(fishFiles, filepath.Join(xdgData, "fish", "fish_history"))
		}
		fishFiles = append(fishFiles, filepath.Join(home, ".local", "share", "fish", "fish_history"))
	}

	if appData := os.Getenv("APPDATA"); appData != "" {
		psFiles = []string{
			filepath.Join(appData, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt"),
		}
	}

	addList := func(list []string) {
		for _, p := range list {
			addFile(p)
		}
	}

	// Prioritize based on current $SHELL
	switch {
	case strings.Contains(shell, "zsh"):
		addList(zshFiles)
		addList(bashFiles)
		addList(fishFiles)
		addList(psFiles)
	case strings.Contains(shell, "fish"):
		addList(fishFiles)
		addList(zshFiles)
		addList(bashFiles)
		addList(psFiles)
	case strings.Contains(shell, "pwsh") || strings.Contains(shell, "powershell"):
		addList(psFiles)
		addList(bashFiles)
		addList(zshFiles)
		addList(fishFiles)
	default: // bash, sh, or unspecified
		addList(bashFiles)
		addList(zshFiles)
		addList(fishFiles)
		addList(psFiles)
	}

	return files
}

// CleanHistoryLine strips line numbers, timestamp metadata, and shell-specific prefixes.
func CleanHistoryLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}

	// Strip leading line numbers if present (e.g. "  493  git status" or "10: git status")
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i > 0 && i < len(line) && (line[i] == ' ' || line[i] == '\t' || line[i] == ':') {
		line = strings.TrimSpace(line[i+1:])
	}

	// Bash timestamp lines: '#1695462000'
	if strings.HasPrefix(line, "#") && len(line) > 1 {
		isTimestamp := true
		for _, r := range line[1:] {
			if r < '0' || r > '9' {
				isTimestamp = false
				break
			}
		}
		if isTimestamp {
			return ""
		}
	}

	// Zsh extended history format: ': 1695462000:0;actual command'
	if strings.HasPrefix(line, ": ") {
		if semi := strings.Index(line, ";"); semi != -1 {
			line = strings.TrimSpace(line[semi+1:])
		}
	}

	// Fish history format: '- cmd: actual command'
	if strings.HasPrefix(line, "- cmd: ") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "- cmd: "))
	} else if strings.HasPrefix(line, "when: ") || strings.HasPrefix(line, "  when: ") ||
		strings.HasPrefix(line, "paths: ") || strings.HasPrefix(line, "  paths: ") {
		return ""
	}

	return line
}

// ReadHistoryFile reads up to maxLines from a history file and returns unique commands (newest first).
func ReadHistoryFile(path string, maxLines int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var rawLines []string
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		rawLines = append(rawLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil && len(rawLines) == 0 {
		return nil, err
	}

	// Process backwards (newest commands at the bottom)
	var commands []string
	seen := make(map[string]bool)

	for i := len(rawLines) - 1; i >= 0 && len(commands) < maxLines; i-- {
		cmd := CleanHistoryLine(rawLines[i])
		if cmd == "" || len(cmd) < 2 || seen[cmd] {
			continue
		}
		seen[cmd] = true
		commands = append(commands, cmd)
	}

	return commands, nil
}

// LoadShellHistory searches default history files and returns recent unique commands.
func LoadShellHistory(limit int) []string {
	if limit <= 0 {
		limit = 500
	}

	var results []string
	seen := make(map[string]bool)

	for _, path := range DefaultHistoryFiles() {
		if fi, err := os.Stat(path); err != nil || fi.IsDir() {
			continue
		}

		cmds, err := ReadHistoryFile(path, limit)
		if err != nil {
			continue
		}

		for _, cmd := range cmds {
			if !seen[cmd] {
				seen[cmd] = true
				results = append(results, cmd)
				if len(results) >= limit {
					return results
				}
			}
		}
	}

	return results
}

// GetSuggestions returns unique commands from shell history.
func (s *Store) GetSuggestions(limit int) []string {
	return LoadShellHistory(limit)
}
