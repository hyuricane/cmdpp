package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModalDropdownFilter(t *testing.T) {
	suggestions := []string{
		"docker compose up -d",
		"docker ps --all",
		"git checkout -b main",
		"kubectl get pods -A",
	}

	modal := NewModal(ModalModeAdd, "", "", "", suggestions)

	// Focus on Command input
	modal.FocusIndex = 1
	modal.Inputs[0].Blur()
	modal.Inputs[1].Focus()

	// Initial view should show recent suggestions
	if len(modal.FilteredSuggestions) != 4 {
		t.Fatalf("expected 4 initial suggestions, got %d", len(modal.FilteredSuggestions))
	}

	// Type "dock"
	for _, r := range "dock" {
		_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if len(modal.FilteredSuggestions) != 2 {
		t.Fatalf("expected 2 filtered suggestions for 'dock', got %d: %v", len(modal.FilteredSuggestions), modal.FilteredSuggestions)
	}
	if modal.FilteredSuggestions[0] != "docker compose up -d" {
		t.Errorf("expected first suggestion 'docker compose up -d', got %q", modal.FilteredSuggestions[0])
	}
	if modal.FilteredSuggestions[1] != "docker ps --all" {
		t.Errorf("expected second suggestion 'docker ps --all', got %q", modal.FilteredSuggestions[1])
	}

	// Check that dropdown view renders
	view := modal.View()
	if !strings.Contains(view, "docker compose up -d") {
		t.Errorf("expected modal view to render dropdown with 'docker compose up -d'")
	}
	if !strings.Contains(view, "Suggestions") {
		t.Errorf("expected modal view to contain 'Suggestions'")
	}
}

func TestModalDropdownNavigationAndPick(t *testing.T) {
	suggestions := []string{
		"npm run build",
		"npm run test",
		"npm run dev",
	}

	modal := NewModal(ModalModeAdd, "", "", "", suggestions)
	modal.FocusIndex = 1
	modal.Inputs[0].Blur()
	modal.Inputs[1].Focus()

	for _, r := range "npm" {
		_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if modal.SelectedSugIndex != -1 {
		t.Fatalf("expected SelectedSugIndex to start at -1, got %d", modal.SelectedSugIndex)
	}

	// 1. Press Down to highlight first item
	_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyDown})
	if modal.SelectedSugIndex != 0 {
		t.Errorf("expected SelectedSugIndex=0 after Down, got %d", modal.SelectedSugIndex)
	}

	// 2. Press Down again to highlight second item
	_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyDown})
	if modal.SelectedSugIndex != 1 {
		t.Errorf("expected SelectedSugIndex=1 after Down, got %d", modal.SelectedSugIndex)
	}

	// 3. Press Up to go back to first item
	_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyUp})
	if modal.SelectedSugIndex != 0 {
		t.Errorf("expected SelectedSugIndex=0 after Up, got %d", modal.SelectedSugIndex)
	}

	// 4. Press Enter to pick item #0 ("npm run build")
	_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if modal.Inputs[1].Value() != "npm run build" {
		t.Errorf("expected command input to be 'npm run build', got %q", modal.Inputs[1].Value())
	}
	if modal.SelectedSugIndex != -1 {
		t.Errorf("expected SelectedSugIndex reset to -1, got %d", modal.SelectedSugIndex)
	}
	// Focus should stay on Command input so user can edit if needed
	if modal.FocusIndex != 1 {
		t.Errorf("expected focus to stay on command input (1), got %d", modal.FocusIndex)
	}
}

func TestModalDropdownTabPickAndNext(t *testing.T) {
	suggestions := []string{
		"ssh -N -L user@remote.org",
	}

	modal := NewModal(ModalModeAdd, "", "", "", suggestions)
	modal.FocusIndex = 1
	modal.Inputs[0].Blur()
	modal.Inputs[1].Focus()

	// Press Down to highlight the suggestion
	_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyDown})
	if modal.SelectedSugIndex != 0 {
		t.Fatalf("expected SelectedSugIndex=0, got %d", modal.SelectedSugIndex)
	}

	// Press Tab to pick AND advance to Description
	_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if modal.Inputs[1].Value() != "ssh -N -L user@remote.org" {
		t.Errorf("expected command input 'ssh -N -L user@remote.org', got %q", modal.Inputs[1].Value())
	}
	if modal.FocusIndex != 2 {
		t.Errorf("expected focus to advance to Description (2), got %d", modal.FocusIndex)
	}
}

func TestModalDropdownEscDismissesSelection(t *testing.T) {
	suggestions := []string{"curl https://example.com"}

	modal := NewModal(ModalModeAdd, "", "", "", suggestions)
	modal.FocusIndex = 1
	modal.Inputs[0].Blur()
	modal.Inputs[1].Focus()

	// Press Down to highlight
	_, _, _ = modal.Update(tea.KeyMsg{Type: tea.KeyDown})
	if modal.SelectedSugIndex != 0 {
		t.Fatalf("expected SelectedSugIndex=0")
	}

	// Press Esc: should deselect dropdown item but NOT close modal
	_, saved, canceled := modal.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if saved || canceled {
		t.Fatalf("expected modal to remain open on first Esc")
	}
	if modal.SelectedSugIndex != -1 {
		t.Errorf("expected SelectedSugIndex to be reset to -1, got %d", modal.SelectedSugIndex)
	}

	// Second Esc closes/cancels modal
	_, _, canceled = modal.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !canceled {
		t.Errorf("expected second Esc to cancel modal")
	}
}
