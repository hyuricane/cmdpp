package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		{"history command with number", "  493  git status", "git status"},
		{"history command with colon", "10: docker ps", "docker ps"},
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

func TestGetSuggestionsPureHistory(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "commands.json")
	histFile := filepath.Join(tempDir, "history.txt")

	t.Setenv("HISTFILE", histFile)

	content := `curl https://api.site.com
docker compose up -d
`
	if err := os.WriteFile(histFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write history file: %v", err)
	}

	s, err := LoadFrom(storePath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	// Add stored commands that should NOT be in suggestions
	_, _, _ = s.AddOrUpdate("c1", "internal-stored-cmd", "desc")

	suggestions := s.GetSuggestions(10)

	for _, sug := range suggestions {
		if sug == "internal-stored-cmd" {
			t.Errorf("expected GetSuggestions to not include stored command 'internal-stored-cmd'")
		}
	}

	foundHistory := false
	for _, sug := range suggestions {
		if sug == "docker compose up -d" {
			foundHistory = true
			break
		}
	}
	if !foundHistory {
		t.Errorf("expected to find 'docker compose up -d' in suggestions, got: %v", suggestions)
	}
}

func TestDefaultHistoryFilesOrder(t *testing.T) {
	t.Setenv("HISTFILE", "")

	// 1. Zsh
	t.Setenv("SHELL", "/bin/zsh")
	zshOrder := DefaultHistoryFiles()
	if len(zshOrder) == 0 || !strings.Contains(zshOrder[0], "zsh") {
		t.Errorf("expected first history file for zsh to be zsh history, got: %v", zshOrder)
	}

	// 2. Fish
	t.Setenv("SHELL", "/usr/bin/fish")
	fishOrder := DefaultHistoryFiles()
	if len(fishOrder) == 0 || !strings.Contains(fishOrder[0], "fish") {
		t.Errorf("expected first history file for fish to be fish history, got: %v", fishOrder)
	}

	// 3. Bash
	t.Setenv("SHELL", "/bin/bash")
	bashOrder := DefaultHistoryFiles()
	if len(bashOrder) == 0 || !strings.Contains(bashOrder[0], "bash") {
		t.Errorf("expected first history file for bash to be bash history, got: %v", bashOrder)
	}
}
