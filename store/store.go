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

// DirPerm defines the secure private directory permissions (0700: rwx------).
const DirPerm os.FileMode = 0700

// FilePerm defines the secure private file permissions (0600: rw-------).
const FilePerm os.FileMode = 0600

// isSystemOrSharedDir returns true if dir is a root, shared, or parent directory
// that should not have its permissions restricted.
func isSystemOrSharedDir(dir string) bool {
	cleanDir := filepath.Clean(dir)
	if cleanDir == "/" || cleanDir == "." || cleanDir == "" {
		return true
	}
	if cleanDir == filepath.Clean(os.TempDir()) || cleanDir == "/tmp" || cleanDir == "/var/tmp" {
		return true
	}
	if home, err := os.UserHomeDir(); err == nil {
		if cleanDir == filepath.Clean(home) || cleanDir == filepath.Join(filepath.Clean(home), ".config") {
			return true
		}
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" && cleanDir == filepath.Clean(xdg) {
		return true
	}
	return false
}

// MigratePermissions verifies and restricts permissions on the directory (mode 0700)
// and store file (mode 0600) to ensure stored commands and tokens are private to the user.
func MigratePermissions(filePath string) error {
	dir := filepath.Dir(filePath)
	if !isSystemOrSharedDir(dir) {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			if fi.Mode().Perm() != DirPerm {
				if err := os.Chmod(dir, DirPerm); err != nil {
					return fmt.Errorf("failed to restrict directory permissions on %s: %w", dir, err)
				}
			}
		}
	}

	if fi, err := os.Stat(filePath); err == nil && !fi.IsDir() {
		if fi.Mode().Perm() != FilePerm {
			if err := os.Chmod(filePath, FilePerm); err != nil {
				return fmt.Errorf("failed to restrict file permissions on %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// LoadFrom loads commands from a specified file path.
// If the file does not exist, an empty store is returned.
// Automatically migrates existing directory and file permissions to private modes (0700/0600).
func LoadFrom(path string) (*Store, error) {
	s := &Store{
		filePath: path,
		commands: make([]Command, 0),
	}

	if err := MigratePermissions(path); err != nil {
		return nil, err
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
// Ensures private permissions (0700 for directory, 0600 for file).
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, DirPerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Ensure directory permissions are private even if it already existed with open permissions
	if err := MigratePermissions(s.filePath); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.commands, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal commands: %w", err)
	}

	// Write to temporary file first for atomic replacement
	tempFile := fmt.Sprintf("%s.tmp.%d", s.filePath, time.Now().UnixNano())
	if err := os.WriteFile(tempFile, data, FilePerm); err != nil {
		return fmt.Errorf("failed to write temporary store file: %w", err)
	}

	// Explicitly chmod the temporary file in case the process umask altered it
	if err := os.Chmod(tempFile, FilePerm); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to set permissions on temporary store file: %w", err)
	}

	if err := os.Rename(tempFile, s.filePath); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("failed to commit store file: %w", err)
	}

	return nil
}

// MigratePermissions checks and restricts permissions on the directory and store file to private modes.
func (s *Store) MigratePermissions() error {
	return MigratePermissions(s.filePath)
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
