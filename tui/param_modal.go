package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hyuricane/cmdpp/store"
)

// ParamModalModel represents the interactive modal dialog for providing command parameter values.
type ParamModalModel struct {
	Command    *store.Command
	Params     []store.Param
	Inputs     []textinput.Model
	FocusIndex int
	ErrMessage string
	Width      int
	Submitted  bool
	Canceled   bool
}

// NewParamModal creates an initialized parameter input modal.
func NewParamModal(cmd *store.Command, params []store.Param) ParamModalModel {
	inputs := make([]textinput.Model, len(params))
	for i, p := range params {
		ti := textinput.New()
		ti.CharLimit = 256
		ti.Width = 46
		if p.HasDefault {
			ti.Placeholder = p.DefaultValue
		} else {
			ti.Placeholder = "required"
		}
		if i == 0 {
			ti.Focus()
		}
		inputs[i] = ti
	}

	return ParamModalModel{
		Command:    cmd,
		Params:     params,
		Inputs:     inputs,
		FocusIndex: 0,
		Width:      64,
	}
}

// Update handles messages and keybindings for the parameter modal.
func (m ParamModalModel) Update(msg tea.Msg) (ParamModalModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Canceled = true
			return m, nil

		case "ctrl+c":
			m.Canceled = true
			return m, tea.Quit

		case "tab", "down":
			if len(m.Inputs) > 1 {
				m.Inputs[m.FocusIndex].Blur()
				m.FocusIndex = (m.FocusIndex + 1) % len(m.Inputs)
				m.Inputs[m.FocusIndex].Focus()
			}
			return m, nil

		case "shift+tab", "up":
			if len(m.Inputs) > 1 {
				m.Inputs[m.FocusIndex].Blur()
				m.FocusIndex = (m.FocusIndex - 1 + len(m.Inputs)) % len(m.Inputs)
				m.Inputs[m.FocusIndex].Focus()
			}
			return m, nil

		case "enter":
			// Validate current field if required
			currParam := m.Params[m.FocusIndex]
			currVal := strings.TrimSpace(m.Inputs[m.FocusIndex].Value())
			if currVal == "" && !currParam.HasDefault {
				m.ErrMessage = fmt.Sprintf("%s is required", currParam.Name)
				return m, nil
			}
			m.ErrMessage = ""

			// If not on the last field, move to the next field
			if m.FocusIndex < len(m.Inputs)-1 {
				m.Inputs[m.FocusIndex].Blur()
				m.FocusIndex++
				m.Inputs[m.FocusIndex].Focus()
				return m, nil
			}

			// On the last field: validate all fields and submit
			for i, p := range m.Params {
				v := strings.TrimSpace(m.Inputs[i].Value())
				if v == "" && !p.HasDefault {
					m.Inputs[m.FocusIndex].Blur()
					m.FocusIndex = i
					m.Inputs[m.FocusIndex].Focus()
					m.ErrMessage = fmt.Sprintf("%s is required", p.Name)
					return m, nil
				}
			}

			m.Submitted = true
			return m, nil
		}
	}

	// Forward other messages (e.g. typing) to the focused text input
	var cmd tea.Cmd
	if m.FocusIndex >= 0 && m.FocusIndex < len(m.Inputs) {
		m.Inputs[m.FocusIndex], cmd = m.Inputs[m.FocusIndex].Update(msg)
	}
	return m, cmd
}

// Values returns the resolved parameter values.
// If an input is empty and the parameter has a default, the default value is returned.
func (m ParamModalModel) Values() map[string]string {
	values := make(map[string]string, len(m.Params))
	for i, p := range m.Params {
		val := strings.TrimSpace(m.Inputs[i].Value())
		if val == "" && p.HasDefault {
			val = p.DefaultValue
		}
		values[p.Name] = val
	}
	return values
}

// Preview returns the command string with currently filled/defaulted parameter values substituted.
func (m ParamModalModel) Preview() string {
	values := m.Values()
	return store.SubstituteParams(m.Command.Cmd, values)
}

// View renders the parameter modal dialog.
func (m ParamModalModel) View() string {
	var b strings.Builder

	// Header / Title
	title := ModalTitleStyle.Render(fmt.Sprintf("Parameters: %s", m.Command.Name))
	b.WriteString(title + "\n\n")

	// Render parameter inputs
	for i, p := range m.Params {
		isFocused := (i == m.FocusIndex)

		var label string
		if p.HasDefault {
			label = fmt.Sprintf("%s (default: %s)", p.Name, p.DefaultValue)
		} else {
			label = fmt.Sprintf("%s *", p.Name)
		}

		labelRendered := ModalFieldLabelStyle.Render(label)
		b.WriteString(labelRendered + "\n")

		inputBoxStyle := ModalInputInactive
		if isFocused {
			inputBoxStyle = ModalInputActive
		}
		b.WriteString(inputBoxStyle.Render(m.Inputs[i].View()) + "\n\n")
	}

	// Live preview
	previewLabel := lipgloss.NewStyle().Bold(true).Foreground(PrimaryColor).Render("Preview:")
	previewCmd := lipgloss.NewStyle().Foreground(TextColor).Italic(true).Render(m.Preview())
	b.WriteString(previewLabel + "\n" + previewCmd + "\n\n")

	// Error message
	if m.ErrMessage != "" {
		b.WriteString(ToastDangerStyle.Render("⚠️  "+m.ErrMessage) + "\n\n")
	}

	// Footer / Keybindings
	var helpBar string
	if len(m.Inputs) == 1 {
		helpBar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			HelpKeyStyle.Render("[Enter]"), HelpDescStyle.Render(" Run   "),
			HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Cancel"),
		)
	} else if m.FocusIndex == len(m.Inputs)-1 {
		helpBar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			HelpKeyStyle.Render("[Enter]"), HelpDescStyle.Render(" Run   "),
			HelpKeyStyle.Render("[Tab]"), HelpDescStyle.Render(" Next field   "),
			HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Cancel"),
		)
	} else {
		helpBar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			HelpKeyStyle.Render("[Enter]"), HelpDescStyle.Render(" Next   "),
			HelpKeyStyle.Render("[Tab]"), HelpDescStyle.Render(" Next field   "),
			HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Cancel"),
		)
	}
	b.WriteString(helpBar)

	boxWidth := max(56, min(m.Width-4, 76))
	return ModalBoxStyle.Width(boxWidth).Render(b.String())
}
