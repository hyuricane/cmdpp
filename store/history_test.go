package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCleanHistoryLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"simple bash", "git status", "git status"},
		{"bash with spaces", "   npm run build   ", "npm run build"},
		{"bash timestamp comment", "#1695462000", ""},
		{"bash comment not timestamp", "# this is a comment", "# this is a comment"},
		{"zsh extended", ": 1695462000:0;docker compose up -d", "docker compose up -d"},
		{"fish cmd", "- cmd: kubectl get pods", "kubectl get pods"},
		{"fish meta when", "  when: 1695462000", ""},
		{"fish meta paths", "  paths: /usr/bin", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanHistoryLine(tt.input)
			if got != tt.expected {
				t.Errorf("CleanHistoryLine(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestReadHistoryFile(t *testing.T) {
	tempDir := t.TempDir()
	histFile := filepath.Join(tempDir, "history.txt")

	content := `git status
ls -la
#1695462000
git status
docker ps
: 1695462010:0;curl https://example.com
`
	if err := os.WriteFile(histFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test history: %v", err)
	}

	cmds, err := ReadHistoryFile(histFile, 10)
	if err != nil {
		t.Fatalf("ReadHistoryFile failed: %v", err)
	}

	// Should be newest first, deduplicated
	expected := []string{
		"curl https://example.com",
		"docker ps",
		"git status",
		"ls -la",
	}

	if len(cmds) != len(expected) {
		t.Fatalf("expected %d commands, got %d: %v", len(expected), len(cmds), cmds)
	}

	for i, exp := range expected {
		if cmds[i] != exp {
			t.Errorf("cmds[%d] = %q, want %q", i, cmds[i], exp)
		}
	}
}

func TestGetSuggestions(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "commands.json")

	s, err := LoadFrom(storePath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	now := time.Now()
	older := now.Add(-1 * time.Hour)

	// Add commands to store
	c1, _, _ := s.AddOrUpdate("c1", "cmd-one", "desc 1")
	c1.LastRunAt = &older
	c1.RunCount = 5

	c2, _, _ := s.AddOrUpdate("c2", "cmd-two", "desc 2")
	c2.LastRunAt = &now
	c2.RunCount = 2

	c3, _, _ := s.AddOrUpdate("c3", "cmd-three", "desc 3")
	c3.RunCount = 10

	s.commands = []Command{*c1, *c2, *c3}

	suggestions := s.GetSuggestions(10)
	if len(suggestions) < 3 {
		t.Fatalf("expected at least 3 suggestions, got %d", len(suggestions))
	}

	// Most recently run (cmd-two) should be first, then cmd-one (run earlier), then cmd-three (never run)
	if suggestions[0] != "cmd-two" {
		t.Errorf("expected suggestions[0] to be cmd-two, got %q", suggestions[0])
	}
	if suggestions[1] != "cmd-one" {
		t.Errorf("expected suggestions[1] to be cmd-one, got %q", suggestions[1])
	}
	if suggestions[2] != "cmd-three" {
		t.Errorf("expected suggestions[2] to be cmd-three, got %q", suggestions[2])
	}
}
