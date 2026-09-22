package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ModalMode int

const (
	ModalModeAdd ModalMode = iota
	ModalModeEdit
)

// ModalModel represents the add/edit command dialog.
type ModalModel struct {
	Mode         ModalMode
	OriginalName string
	FocusIndex   int
	Inputs       []textinput.Model
	ErrMessage   string
	Width        int
}

// NewModal creates an initialized modal dialog.
func NewModal(mode ModalMode, initialName, initialCmd, initialDesc string) ModalModel {
	inputs := make([]textinput.Model, 3)

	// 1. Name input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "tunnel-one"
	inputs[0].Focus()
	inputs[0].CharLimit = 64
	inputs[0].Width = 50
	inputs[0].SetValue(initialName)

	// 2. Command input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "ssh -N -L user@ssh.example.org"
	inputs[1].CharLimit = 1024
	inputs[1].Width = 50
	inputs[1].SetValue(initialCmd)

	// 3. Description input
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Optional description or notes"
	inputs[2].CharLimit = 256
	inputs[2].Width = 50
	inputs[2].SetValue(initialDesc)

	focusIdx := 0
	if mode == ModalModeEdit {
		// When editing, start focused on the command string for convenience
		inputs[0].Blur()
		inputs[1].Focus()
		focusIdx = 1
	}

	return ModalModel{
		Mode:         mode,
		OriginalName: initialName,
		FocusIndex:   focusIdx,
		Inputs:       inputs,
		Width:        60,
	}
}

// Update handles key presses in the modal.
func (m *ModalModel) Update(msg tea.Msg) (tea.Cmd, bool, bool) {
	// returns: (cmd, saved, canceled)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return nil, false, true

		case "tab", "down":
			m.nextFocus()
			return nil, false, false

		case "shift+tab", "up":
			m.prevFocus()
			return nil, false, false

		case "ctrl+s":
			if m.validate() {
				return nil, true, false
			}
			return nil, false, false

		case "enter":
			if m.FocusIndex < len(m.Inputs)-1 {
				m.nextFocus()
				return nil, false, false
			}
			if m.validate() {
				return nil, true, false
			}
			return nil, false, false
		}
	}

	var cmds []tea.Cmd
	for i := range m.Inputs {
		var cmd tea.Cmd
		m.Inputs[i], cmd = m.Inputs[i].Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...), false, false
}

func (m *ModalModel) nextFocus() {
	m.Inputs[m.FocusIndex].Blur()
	m.FocusIndex = (m.FocusIndex + 1) % len(m.Inputs)
	m.Inputs[m.FocusIndex].Focus()
}

func (m *ModalModel) prevFocus() {
	m.Inputs[m.FocusIndex].Blur()
	m.FocusIndex = (m.FocusIndex - 1 + len(m.Inputs)) % len(m.Inputs)
	m.Inputs[m.FocusIndex].Focus()
}

func (m *ModalModel) validate() bool {
	name := strings.TrimSpace(m.Inputs[0].Value())
	cmd := strings.TrimSpace(m.Inputs[1].Value())

	if name == "" {
		m.ErrMessage = "Command name cannot be empty"
		m.Inputs[0].Focus()
		m.FocusIndex = 0
		return false
	}

	if cmd == "" {
		m.ErrMessage = "Command string cannot be empty"
		m.Inputs[1].Focus()
		m.FocusIndex = 1
		return false
	}

	m.ErrMessage = ""
	return true
}

// Values returns the sanitized name, command string, and description.
func (m *ModalModel) Values() (string, string, string) {
	return strings.TrimSpace(m.Inputs[0].Value()),
		strings.TrimSpace(m.Inputs[1].Value()),
		strings.TrimSpace(m.Inputs[2].Value())
}

// View renders the modal dialog.
func (m *ModalModel) View() string {
	var title string
	if m.Mode == ModalModeAdd {
		title = ModalTitleStyle.Render("➕ Add New Command")
	} else {
		title = ModalTitleStyle.Render(fmt.Sprintf("✏️ Edit Command: %s", m.OriginalName))
	}

	var b strings.Builder
	b.WriteString(title + "\n\n")

	labels := []string{"Name:", "Command:", "Description:"}
	for i, input := range m.Inputs {
		label := ModalFieldLabelStyle.Render(labels[i])
		boxStyle := ModalInputInactive
		if i == m.FocusIndex {
			boxStyle = ModalInputActive
		}
		b.WriteString(label + "\n")
		b.WriteString(boxStyle.Render(input.View()) + "\n\n")
	}

	if m.ErrMessage != "" {
		b.WriteString(ToastDangerStyle.Render("⚠️ "+m.ErrMessage) + "\n\n")
	}

	controls := lipgloss.JoinHorizontal(
		lipgloss.Left,
		HelpKeyStyle.Render("[Tab]"), HelpDescStyle.Render(" Next   "),
		HelpKeyStyle.Render("[Enter/Ctrl+S]"), HelpDescStyle.Render(" Save   "),
		HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Cancel"),
	)
	b.WriteString(controls)

	return ModalBoxStyle.Render(b.String())
}
