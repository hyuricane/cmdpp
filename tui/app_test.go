package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"cmdpp/store"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIFlow(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "commands.json")

	s, err := store.LoadFrom(storePath)
	if err != nil {
		t.Fatalf("LoadFrom failed: %v", err)
	}

	_, _, _ = s.AddOrUpdate("tunnel-one", "ssh -N -L user@ssh.example.org", "Tunnel 1")
	_, _, _ = s.AddOrUpdate("docker-ps", "docker ps --format 'table {{.Names}}'", "List containers")
	_ = s.Save()

	m := NewModel(s)
	// Send window size
	modelInterface, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = modelInterface.(Model)

	// 1. Check initial view rendering
	view := m.View()
	if !strings.Contains(view, "cmdpp") {
		t.Errorf("expected view to contain 'cmdpp', got:\n%s", view)
	}
	if !strings.Contains(view, "tunnel-one") {
		t.Errorf("expected view to contain 'tunnel-one'")
	}
	if !strings.Contains(view, "docker-ps") {
		t.Errorf("expected view to contain 'docker-ps'")
	}

	// 2. Navigation: Down arrow
	modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = modelInterface.(Model)
	if m.cursor != 1 {
		t.Errorf("expected cursor=1, got %d", m.cursor)
	}

	// 3. Selection / Run: Enter
	modelInterface, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = modelInterface.(Model)
	if m.Action() != ActionRun {
		t.Errorf("expected ActionRun, got %v", m.Action())
	}
	if m.ChosenCommand() == nil || m.ChosenCommand().Name != m.filtered[1].Name {
		t.Errorf("expected chosen command %s, got %v", m.filtered[1].Name, m.ChosenCommand())
	}
	if cmd == nil {
		t.Errorf("expected quit cmd, got nil")
	}

	// 4. Test Filtering
	m = NewModel(s)
	modelInterface, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	m = modelInterface.(Model)

	// Press '/' to focus search
	modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = modelInterface.(Model)
	if !m.searchFocused {
		t.Errorf("expected searchFocused=true")
	}

	// Type 'docker'
	for _, r := range "docker" {
		modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = modelInterface.(Model)
	}
	if len(m.filtered) != 1 || m.filtered[0].Name != "docker-ps" {
		t.Errorf("expected 1 filtered item (docker-ps), got %d items", len(m.filtered))
	}

	// Press Esc to exit search
	modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = modelInterface.(Model)
	if m.searchFocused {
		t.Errorf("expected searchFocused=false after Esc")
	}

	// 5. Test Add Modal
	modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = modelInterface.(Model)
	if m.modal == nil || m.modal.Mode != ModalModeAdd {
		t.Fatalf("expected Add modal to be active")
	}

	// Set inputs in modal
	m.modal.Inputs[0].SetValue("k8s-pods")
	m.modal.Inputs[1].SetValue("kubectl get pods -A")
	m.modal.Inputs[2].SetValue("List all Kubernetes pods")

	// Submit modal with Enter
	m.modal.FocusIndex = 2
	modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = modelInterface.(Model)
	if m.modal != nil {
		t.Errorf("expected modal to close after saving")
	}
	if _, ok := s.Get("k8s-pods"); !ok {
		t.Errorf("expected 'k8s-pods' to be saved in store")
	}

	// 6. Test Delete confirmation
	m.cursor = 0
	targetName := m.filtered[0].Name
	modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = modelInterface.(Model)
	if !m.isConfirmingDel {
		t.Errorf("expected isConfirmingDel=true")
	}

	// Confirm delete with 'y'
	modelInterface, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = modelInterface.(Model)
	if _, ok := s.Get(targetName); ok {
		t.Errorf("expected %q to be deleted from store", targetName)
	}
}
