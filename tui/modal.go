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

const maxVisibleDropdownItems = 4

// ModalModel represents the add/edit command dialog.
type ModalModel struct {
	Mode                ModalMode
	OriginalName        string
	FocusIndex          int
	Inputs              []textinput.Model
	ErrMessage          string
	Width               int
	AllSuggestions      []string
	FilteredSuggestions []string
	SelectedSugIndex    int // -1 when typing in the text input, >= 0 when a dropdown item is highlighted
}

// NewModal creates an initialized modal dialog.
// suggestionList optionally provides autocomplete command suggestions (e.g. from shell history).
func NewModal(mode ModalMode, initialName, initialCmd, initialDesc string, suggestionList ...[]string) ModalModel {
	var suggestions []string
	if len(suggestionList) > 0 {
		suggestions = suggestionList[0]
	}

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

	m := ModalModel{
		Mode:             mode,
		OriginalName:     initialName,
		FocusIndex:       focusIdx,
		Inputs:           inputs,
		Width:            60,
		AllSuggestions:   suggestions,
		SelectedSugIndex: -1,
	}

	m.updateSuggestions()
	return m
}

func (m *ModalModel) updateSuggestions() {
	if len(m.AllSuggestions) == 0 {
		m.FilteredSuggestions = nil
		m.SelectedSugIndex = -1
		return
	}

	query := strings.TrimSpace(m.Inputs[1].Value())
	if query == "" {
		limit := min(maxVisibleDropdownItems, len(m.AllSuggestions))
		m.FilteredSuggestions = make([]string, limit)
		copy(m.FilteredSuggestions, m.AllSuggestions[:limit])
		m.SelectedSugIndex = -1
		return
	}

	queryLower := strings.ToLower(query)
	var prefixMatches []string
	var containsMatches []string

	for _, s := range m.AllSuggestions {
		sLower := strings.ToLower(s)
		if strings.HasPrefix(sLower, queryLower) {
			prefixMatches = append(prefixMatches, s)
		} else if strings.Contains(sLower, queryLower) {
			containsMatches = append(containsMatches, s)
		}
		if len(prefixMatches)+len(containsMatches) >= maxVisibleDropdownItems*4 {
			break
		}
	}

	matches := append(prefixMatches, containsMatches...)
	if len(matches) > maxVisibleDropdownItems {
		matches = matches[:maxVisibleDropdownItems]
	}

	// If the only matching suggestion is identical to what's already typed, hide dropdown
	if len(matches) == 1 && matches[0] == strings.TrimSpace(m.Inputs[1].Value()) {
		m.FilteredSuggestions = nil
		m.SelectedSugIndex = -1
		return
	}

	m.FilteredSuggestions = matches
	if m.SelectedSugIndex >= len(m.FilteredSuggestions) {
		m.SelectedSugIndex = len(m.FilteredSuggestions) - 1
	}
}

// Update handles key presses in the modal.
func (m *ModalModel) Update(msg tea.Msg) (tea.Cmd, bool, bool) {
	// returns: (cmd, saved, canceled)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.FocusIndex == 1 && m.SelectedSugIndex >= 0 {
				m.SelectedSugIndex = -1
				return nil, false, false
			}
			return nil, false, true

		case "tab":
			if m.FocusIndex == 1 && m.SelectedSugIndex >= 0 && m.SelectedSugIndex < len(m.FilteredSuggestions) {
				m.Inputs[1].SetValue(m.FilteredSuggestions[m.SelectedSugIndex])
				m.Inputs[1].CursorEnd()
				m.SelectedSugIndex = -1
				m.updateSuggestions()
				m.nextFocus()
				return nil, false, false
			}
			m.nextFocus()
			return nil, false, false

		case "shift+tab":
			m.prevFocus()
			return nil, false, false

		case "down", "ctrl+n":
			if m.FocusIndex == 1 && len(m.FilteredSuggestions) > 0 {
				if m.SelectedSugIndex < len(m.FilteredSuggestions)-1 {
					m.SelectedSugIndex++
					return nil, false, false
				}
				// Reached bottom of dropdown: advance to next field
				m.SelectedSugIndex = -1
				m.nextFocus()
				return nil, false, false
			}
			m.nextFocus()
			return nil, false, false

		case "up", "ctrl+p":
			if m.FocusIndex == 1 && len(m.FilteredSuggestions) > 0 {
				if m.SelectedSugIndex > 0 {
					m.SelectedSugIndex--
					return nil, false, false
				} else if m.SelectedSugIndex == 0 {
					m.SelectedSugIndex = -1
					return nil, false, false
				}
			}
			m.prevFocus()
			return nil, false, false

		case "enter":
			if m.FocusIndex == 1 && m.SelectedSugIndex >= 0 && m.SelectedSugIndex < len(m.FilteredSuggestions) {
				m.Inputs[1].SetValue(m.FilteredSuggestions[m.SelectedSugIndex])
				m.Inputs[1].CursorEnd()
				m.SelectedSugIndex = -1
				m.updateSuggestions()
				return nil, false, false
			}
			if m.FocusIndex < len(m.Inputs)-1 {
				m.nextFocus()
				return nil, false, false
			}
			if m.validate() {
				return nil, true, false
			}
			return nil, false, false

		case "ctrl+s":
			if m.validate() {
				return nil, true, false
			}
			return nil, false, false
		}
	}

	// Update text inputs
	var cmds []tea.Cmd
	for i := range m.Inputs {
		var cmd tea.Cmd
		prevVal := m.Inputs[i].Value()
		m.Inputs[i], cmd = m.Inputs[i].Update(msg)
		if i == 1 && m.Inputs[1].Value() != prevVal {
			m.SelectedSugIndex = -1
			m.updateSuggestions()
		}
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...), false, false
}

func (m *ModalModel) nextFocus() {
	m.Inputs[m.FocusIndex].Blur()
	m.FocusIndex = (m.FocusIndex + 1) % len(m.Inputs)
	m.Inputs[m.FocusIndex].Focus()
	m.SelectedSugIndex = -1
}

func (m *ModalModel) prevFocus() {
	m.Inputs[m.FocusIndex].Blur()
	m.FocusIndex = (m.FocusIndex - 1 + len(m.Inputs)) % len(m.Inputs)
	m.Inputs[m.FocusIndex].Focus()
	m.SelectedSugIndex = -1
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

func (m *ModalModel) renderDropdown() string {
	var sb strings.Builder
	sb.WriteString(DropdownHeaderStyle.Render("▾ Suggestions (from history):") + "\n")

	maxCmdWidth := max(30, m.Width-16)
	for idx, sug := range m.FilteredSuggestions {
		disp := sug
		if len(disp) > maxCmdWidth {
			disp = disp[:maxCmdWidth-3] + "..."
		}

		if idx == m.SelectedSugIndex {
			line := fmt.Sprintf("▸ %s", disp)
			pad := max(0, maxCmdWidth+2-len(disp))
			line += strings.Repeat(" ", pad)
			sb.WriteString(DropdownItemSelected.Render(line))
		} else {
			line := fmt.Sprintf("  %s", disp)
			sb.WriteString(DropdownItemNormal.Render(line))
		}
		if idx < len(m.FilteredSuggestions)-1 {
			sb.WriteString("\n")
		}
	}

	return DropdownContainerStyle.Width(m.Width - 10).Render(sb.String())
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
		if i == m.FocusIndex && (i != 1 || m.SelectedSugIndex == -1) {
			boxStyle = ModalInputActive
		}
		b.WriteString(label + "\n")
		b.WriteString(boxStyle.Render(input.View()) + "\n")

		if i == 1 && m.FocusIndex == 1 && len(m.FilteredSuggestions) > 0 {
			b.WriteString(m.renderDropdown() + "\n")
		} else {
			b.WriteString("\n")
		}
	}

	if m.ErrMessage != "" {
		b.WriteString(ToastDangerStyle.Render("⚠️ "+m.ErrMessage) + "\n\n")
	}

	var controls string
	if m.FocusIndex == 1 && len(m.FilteredSuggestions) > 0 {
		if m.SelectedSugIndex >= 0 {
			controls = lipgloss.JoinHorizontal(
				lipgloss.Left,
				HelpKeyStyle.Render("[Enter]"), HelpDescStyle.Render(" Pick   "),
				HelpKeyStyle.Render("[Tab]"), HelpDescStyle.Render(" Pick & Next   "),
				HelpKeyStyle.Render("[↑/↓]"), HelpDescStyle.Render(" Navigate   "),
				HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Back to Input"),
			)
		} else {
			controls = lipgloss.JoinHorizontal(
				lipgloss.Left,
				HelpKeyStyle.Render("[↓]"), HelpDescStyle.Render(" Dropdown   "),
				HelpKeyStyle.Render("[Tab]"), HelpDescStyle.Render(" Next   "),
				HelpKeyStyle.Render("[Enter]"), HelpDescStyle.Render(" Save/Next   "),
				HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Cancel"),
			)
		}
	} else {
		controls = lipgloss.JoinHorizontal(
			lipgloss.Left,
			HelpKeyStyle.Render("[Tab]"), HelpDescStyle.Render(" Next   "),
			HelpKeyStyle.Render("[Enter/Ctrl+S]"), HelpDescStyle.Render(" Save   "),
			HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Cancel"),
		)
	}
	b.WriteString(controls)

	return ModalBoxStyle.Render(b.String())
}
