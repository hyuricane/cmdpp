package store

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DefaultHistoryFiles returns candidate shell history paths for the current user and OS.
func DefaultHistoryFiles() []string {
	var files []string

	// 1. Explicit $HISTFILE
	if hf := os.Getenv("HISTFILE"); hf != "" {
		files = append(files, hf)
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		// Bash
		files = append(files, filepath.Join(home, ".bash_history"))
		// Zsh
		files = append(files, filepath.Join(home, ".zsh_history"))
		files = append(files, filepath.Join(home, ".histfile"))
		// Fish
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			files = append(files, filepath.Join(xdgData, "fish", "fish_history"))
		}
		files = append(files, filepath.Join(home, ".local", "share", "fish", "fish_history"))
	}

	// Windows PowerShell PSReadLine
	if appData := os.Getenv("APPDATA"); appData != "" {
		files = append(files, filepath.Join(appData, "Microsoft", "Windows", "PowerShell", "PSReadLine", "ConsoleHost_history.txt"))
	}

	return files
}

// CleanHistoryLine strips timestamp metadata and shell-specific prefixes.
func CleanHistoryLine(line string) string {
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
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

// GetSuggestions returns unique commands combined from store and shell history.
// Stored commands appear first (prioritizing recently and frequently run commands),
// followed by commands from shell history.
func (s *Store) GetSuggestions(limit int) []string {
	if limit <= 0 {
		limit = 500
	}

	var results []string
	seen := make(map[string]bool)

	if s != nil {
		cmds := s.List()
		sort.SliceStable(cmds, func(i, j int) bool {
			if cmds[i].LastRunAt != nil && cmds[j].LastRunAt != nil {
				return cmds[i].LastRunAt.After(*cmds[j].LastRunAt)
			}
			if cmds[i].LastRunAt != nil {
				return true
			}
			if cmds[j].LastRunAt != nil {
				return false
			}
			if cmds[i].RunCount != cmds[j].RunCount {
				return cmds[i].RunCount > cmds[j].RunCount
			}
			return cmds[i].CreatedAt.After(cmds[j].CreatedAt)
		})

		for _, c := range cmds {
			cmdStr := strings.TrimSpace(c.Cmd)
			if cmdStr != "" && !seen[cmdStr] {
				seen[cmdStr] = true
				results = append(results, cmdStr)
				if len(results) >= limit {
					return results
				}
			}
		}
	}

	// Supplement with shell history
	remaining := limit - len(results)
	if remaining > 0 {
		shellCmds := LoadShellHistory(remaining)
		for _, cmd := range shellCmds {
			if !seen[cmd] {
				seen[cmd] = true
				results = append(results, cmd)
				if len(results) >= limit {
					break
				}
			}
		}
	}

	return results
}
