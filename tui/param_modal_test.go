package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hyuricane/cmdpp/store"
)

func TestParamModalSingleParamEnterRuns(t *testing.T) {
	cmd := &store.Command{
		Name: "tunnel",
		Cmd:  "ssh -N -L ${PORT:-3000}:127.0.0.1:${PORT:-3000} user@host.com",
	}
	params := store.ExtractParams(cmd.Cmd)
	modal := NewParamModal(cmd, params)

	if len(modal.Inputs) != 1 {
		t.Fatalf("expected 1 input, got %d", len(modal.Inputs))
	}

	// Pressing Enter immediately on single param with default should submit with default
	updated, _ := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !updated.Submitted {
		t.Errorf("expected modal to be submitted on Enter")
	}

	vals := updated.Values()
	if vals["PORT"] != "3000" {
		t.Errorf("expected PORT to be '3000', got %q", vals["PORT"])
	}

	expectedPreview := "ssh -N -L 3000:127.0.0.1:3000 user@host.com"
	if updated.Preview() != expectedPreview {
		t.Errorf("expected preview %q, got %q", expectedPreview, updated.Preview())
	}
}

func TestParamModalMultiParamEnterNavigation(t *testing.T) {
	cmd := &store.Command{
		Name: "deploy",
		Cmd:  "deploy --host=${HOST} --port=${PORT:-8080}",
	}
	params := store.ExtractParams(cmd.Cmd)
	modal := NewParamModal(cmd, params)

	if len(modal.Inputs) != 2 {
		t.Fatalf("expected 2 inputs, got %d", len(modal.Inputs))
	}

	// 1. First field is required (HOST). Pressing Enter without typing should show error
	m1, _ := modal.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m1.Submitted {
		t.Errorf("should not submit with empty required field")
	}
	if m1.ErrMessage == "" {
		t.Errorf("expected error message for required field")
	}
	if m1.FocusIndex != 0 {
		t.Errorf("focus should stay on field 0, got %d", m1.FocusIndex)
	}

	// 2. Type value for HOST
	m1.Inputs[0].SetValue("example.org")
	m2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Enter should advance focus to field 1 (PORT)
	if m2.Submitted {
		t.Errorf("should advance to next field, not submit yet")
	}
	if m2.FocusIndex != 1 {
		t.Errorf("expected focus to be 1, got %d", m2.FocusIndex)
	}

	// 3. On field 1 (PORT has default), pressing Enter should submit
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m3.Submitted {
		t.Errorf("expected modal to be submitted on last field Enter")
	}

	vals := m3.Values()
	if vals["HOST"] != "example.org" || vals["PORT"] != "8080" {
		t.Errorf("unexpected values: %+v", vals)
	}
}

func TestParamModalTabNavigation(t *testing.T) {
	cmd := &store.Command{
		Name: "test",
		Cmd:  "run ${A:-1} ${B:-2} ${C:-3}",
	}
	params := store.ExtractParams(cmd.Cmd)
	modal := NewParamModal(cmd, params)

	if modal.FocusIndex != 0 {
		t.Fatalf("expected initial focus 0, got %d", modal.FocusIndex)
	}

	// Press Tab
	m1, _ := modal.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m1.FocusIndex != 1 {
		t.Errorf("expected focus 1 after tab, got %d", m1.FocusIndex)
	}

	// Press Tab
	m2, _ := m1.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m2.FocusIndex != 2 {
		t.Errorf("expected focus 2 after tab, got %d", m2.FocusIndex)
	}

	// Press Tab (wraps to 0)
	m3, _ := m2.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m3.FocusIndex != 0 {
		t.Errorf("expected focus 0 after wrap, got %d", m3.FocusIndex)
	}

	// Press Shift+Tab (back to 2)
	m4, _ := m3.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m4.FocusIndex != 2 {
		t.Errorf("expected focus 2 after shift+tab, got %d", m4.FocusIndex)
	}
}

func TestParamModalCancel(t *testing.T) {
	cmd := &store.Command{
		Name: "test",
		Cmd:  "run ${PARAM:-1}",
	}
	params := store.ExtractParams(cmd.Cmd)
	modal := NewParamModal(cmd, params)

	m1, _ := modal.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m1.Canceled {
		t.Errorf("expected Canceled to be true on Esc")
	}
}
