package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Command represents a stored CLI command.
type Command struct {
	Name        string     `json:"name"`
	Cmd         string     `json:"cmd"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	RunCount    int        `json:"run_count"`
}

// Store manages the persistence and retrieval of commands.
type Store struct {
	mu       sync.RWMutex
	filePath string
	commands []Command
}

// ConfigFile returns the path to the commands JSON file.
// Follows $CMDPP_FILE, $XDG_CONFIG_HOME/cmdpp/commands.json, or ~/.config/cmdpp/commands.json.
func ConfigFile() (string, error) {
	if custom := os.Getenv("CMDPP_FILE"); custom != "" {
		return custom, nil
	}

	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to determine home directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	return filepath.Join(configDir, "cmdpp", "commands.json"), nil
}

// Load loads the store from the default configuration path.
func Load() (*Store, error) {
	path, err := ConfigFile()
	if err != nil {
		return nil, err
	}
	return LoadFrom(path)
}

// LoadFrom loads commands from a specified file path.
// If the file does not exist, an empty store is returned.
func LoadFrom(path string) (*Store, error) {
	s := &Store{
		filePath: path,
		commands: make([]Command, 0),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("failed to read commands file: %w", err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return s, nil
	}

	var commands []Command
	if err := json.Unmarshal(data, &commands); err != nil {
		return nil, fmt.Errorf("failed to parse commands file (%s): %w", path, err)
	}

	s.commands = commands
	return s, nil
}

// Save writes commands atomically to the store's file path.
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(s.commands, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	// Write to temporary file first for atomic replacement
	tempFile := fmt.Sprintf("%s.tmp.%d", s.filePath, time.Now().UnixNano())
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary store file: %w", err)
	}

	if err := os.Rename(tempFile, s.filePath); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to commit store file: %w", err)
	}

	return nil
}

// AddOrUpdate adds a command or updates it if one with the same name already exists.
// Returns the saved command, a boolean indicating whether it was an update, and any error.
func (s *Store) AddOrUpdate(name, cmdStr, desc string) (*Command, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.TrimSpace(name)
	cmdStr = strings.TrimSpace(cmdStr)
	desc = strings.TrimSpace(desc)

	if name == "" {
		return nil, false, errors.New("command name cannot be empty")
	}
	if cmdStr == "" {
		return nil, false, errors.New("command string cannot be empty")
	}

	now := time.Now()
	for i, c := range s.commands {
		if strings.EqualFold(c.Name, name) {
			s.commands[i].Cmd = cmdStr
			if desc != "" {
				s.commands[i].Description = desc
			}
			s.commands[i].UpdatedAt = now
			cmd := s.commands[i]
			return &cmd, true, nil
		}
	}

	newCmd := Command{
		Name:        name,
		Cmd:         cmdStr,
		Description: desc,
		CreatedAt:   now,
		UpdatedAt:   now,
		RunCount:    0,
	}
	s.commands = append(s.commands, newCmd)
	return &newCmd, false, nil
}

// Remove deletes a command by name.
func (s *Store) Remove(name string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name = strings.TrimSpace(name)
	for i, c := range s.commands {
		if strings.EqualFold(c.Name, name) {
			s.commands = append(s.commands[:i], s.commands[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// Get finds a command by name.
func (s *Store) Get(name string) (*Command, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	name = strings.TrimSpace(name)
	for _, c := range s.commands {
		if strings.EqualFold(c.Name, name) {
			cp := c
			return &cp, true
		}
	}
	return nil, false
}

// RecordRun updates the run count and last run timestamp for a command.
func (s *Store) RecordRun(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for i, c := range s.commands {
		if strings.EqualFold(c.Name, strings.TrimSpace(name)) {
			s.commands[i].RunCount++
			s.commands[i].LastRunAt = &now
			return nil
		}
	}
	return fmt.Errorf("command %q not found", name)
}

// List returns a copy of all commands.
func (s *Store) List() []Command {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]Command, len(s.commands))
	copy(res, s.commands)
	return res
}

// SortedList returns commands sorted by a specified key: "name", "recent", "runs".
func (s *Store) SortedList(sortBy string) []Command {
	list := s.List()
	switch sortBy {
	case "recent":
		sort.Slice(list, func(i, j int) bool {
			ti := list[i].CreatedAt
			if list[i].LastRunAt != nil {
				ti = *list[i].LastRunAt
			}
			tj := list[j].CreatedAt
			if list[j].LastRunAt != nil {
				tj = *list[j].LastRunAt
			}
			return ti.After(tj)
		})
	case "runs":
		sort.Slice(list, func(i, j int) bool {
			return list[i].RunCount > list[j].RunCount
		})
	default: // name
		sort.Slice(list, func(i, j int) bool {
			return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name)
		})
	}
	return list
}

// FilePath returns the file path of this store.
func (s *Store) FilePath() string {
	return s.filePath
}
