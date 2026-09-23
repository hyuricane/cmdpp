package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hyuricane/cmdpp/cli"
	"github.com/hyuricane/cmdpp/store"
)

type AppAction int

const (
	ActionNone AppAction = iota
	ActionRun
	ActionQuit
)

type clearToastMsg struct{}

// Model is the main Bubble Tea model for cmdpp.
type Model struct {
	store           *store.Store
	commands        []store.Command
	filtered        []store.Command
	cursor          int
	viewportTop     int
	maxVisibleRows  int
	width           int
	height          int
	searchInput     textinput.Model
	searchFocused   bool
	modal           *ModalModel
	isConfirmingDel bool
	toastMsg        string
	toastIsErr      bool
	chosenCommand   *store.Command
	action          AppAction
}

// NewModel initializes the TUI model.
func NewModel(s *store.Store) Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter commands... (press / to search)"
	ti.CharLimit = 100
	ti.Width = 60

	allCmds := s.SortedList("recent")

	m := Model{
		store:          s,
		commands:       allCmds,
		filtered:       allCmds,
		cursor:         0,
		viewportTop:    0,
		maxVisibleRows: 8,
		width:          80,
		height:         24,
		searchInput:    ti,
		searchFocused:  false,
		action:         ActionNone,
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) refreshCommands() {
	m.commands = m.store.SortedList("recent")
	m.filterCommands()
}

func (m *Model) filterCommands() {
	query := strings.TrimSpace(strings.ToLower(m.searchInput.Value()))
	if query == "" {
		m.filtered = m.commands
	} else {
		var matched []store.Command
		for _, c := range m.commands {
			if strings.Contains(strings.ToLower(c.Name), query) ||
				strings.Contains(strings.ToLower(c.Cmd), query) ||
				strings.Contains(strings.ToLower(c.Description), query) {
				matched = append(matched, c)
			}
		}
		m.filtered = matched
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	if m.viewportTop > m.cursor {
		m.viewportTop = m.cursor
	}
}

func (m *Model) setToast(msg string, isErr bool) tea.Cmd {
	m.toastMsg = msg
	m.toastIsErr = isErr
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearToastMsg{}
	})
}

// Update handles events and key presses.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchInput.Width = max(20, msg.Width-14)
		m.maxVisibleRows = max(4, msg.Height-18)
		return m, nil

	case clearToastMsg:
		m.toastMsg = ""
		return m, nil

	case tea.KeyMsg:
		// 1. If Add/Edit Modal is active
		if m.modal != nil {
			cmd, saved, canceled := m.modal.Update(msg)
			if canceled {
				m.modal = nil
				return m, nil
			}
			if saved {
				name, cmdStr, desc := m.modal.Values()
				if m.modal.Mode == ModalModeEdit && m.modal.OriginalName != name {
					// Name changed in edit: remove old one first
					_, _ = m.store.Remove(m.modal.OriginalName)
				}
				savedCmd, _, err := m.store.AddOrUpdate(name, cmdStr, desc)
				if err != nil {
					m.modal.ErrMessage = err.Error()
					return m, nil
				}
				_ = m.store.Save()
				m.modal = nil
				m.refreshCommands()

				// Focus on the newly saved item
				for i, c := range m.filtered {
					if c.Name == savedCmd.Name {
						m.cursor = i
						break
					}
				}
				return m, m.setToast(fmt.Sprintf("✓ Saved command %q", savedCmd.Name), false)
			}
			return m, cmd
		}

		// 2. If confirming deletion
		if m.isConfirmingDel {
			switch msg.String() {
			case "y", "Y":
				if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
					target := m.filtered[m.cursor]
					_, _ = m.store.Remove(target.Name)
					_ = m.store.Save()
					m.isConfirmingDel = false
					m.refreshCommands()
					return m, m.setToast(fmt.Sprintf("✓ Deleted command %q", target.Name), false)
				}
				m.isConfirmingDel = false
				return m, nil
			default:
				m.isConfirmingDel = false
				return m, nil
			}
		}

		// 3. Global hotkeys
		switch msg.String() {
		case "ctrl+c":
			m.action = ActionQuit
			return m, tea.Quit
		}

		// 4. If search bar is focused
		if m.searchFocused {
			switch msg.String() {
			case "esc":
				m.searchFocused = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				// If query has matches, pressing enter can either run selected or blur search
				if len(m.filtered) > 0 {
					sel := m.filtered[m.cursor]
					m.chosenCommand = &sel
					m.action = ActionRun
					return m, tea.Quit
				}
				m.searchFocused = false
				m.searchInput.Blur()
				return m, nil
			case "up":
				m.moveCursorUp()
				return m, nil
			case "down":
				m.moveCursorDown()
				return m, nil
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				m.filterCommands()
				return m, cmd
			}
		}

		// 5. Normal list navigation and command hotkeys
		switch msg.String() {
		case "q":
			m.action = ActionQuit
			return m, tea.Quit

		case "esc":
			if m.searchInput.Value() != "" {
				m.searchInput.SetValue("")
				m.filterCommands()
				return m, nil
			}
			m.action = ActionQuit
			return m, tea.Quit

		case "enter":
			if len(m.filtered) > 0 {
				sel := m.filtered[m.cursor]
				m.chosenCommand = &sel
				m.action = ActionRun
				return m, tea.Quit
			}

		case "/":
			m.searchFocused = true
			m.searchInput.Focus()
			return m, textinput.Blink

		case "up", "k":
			m.moveCursorUp()
			return m, nil

		case "down", "j":
			m.moveCursorDown()
			return m, nil

		case "a":
			modal := NewModal(ModalModeAdd, "", "", "")
			m.modal = &modal
			return m, nil

		case "e":
			if len(m.filtered) > 0 {
				sel := m.filtered[m.cursor]
				modal := NewModal(ModalModeEdit, sel.Name, sel.Cmd, sel.Description)
				m.modal = &modal
				return m, nil
			}

		case "d", "x":
			if len(m.filtered) > 0 {
				m.isConfirmingDel = true
				return m, nil
			}

		case "c":
			if len(m.filtered) > 0 {
				sel := m.filtered[m.cursor]
				if err := CopyToClipboard(sel.Cmd); err != nil {
					return m, m.setToast("Failed to copy to clipboard", true)
				}
				return m, m.setToast(fmt.Sprintf("📋 Copied to clipboard: %s", sel.Cmd), false)
			}
		}
	}

	return m, nil
}

func (m *Model) moveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
		if m.cursor < m.viewportTop {
			m.viewportTop = m.cursor
		}
	}
}

func (m *Model) moveCursorDown() {
	if m.cursor < len(m.filtered)-1 {
		m.cursor++
		if m.cursor >= m.viewportTop+m.maxVisibleRows {
			m.viewportTop = m.cursor - m.maxVisibleRows + 1
		}
	}
}

// View renders the TUI screen.
func (m Model) View() string {
	if m.modal != nil {
		modalView := m.modal.View()
		if m.width > 0 && m.height > 0 {
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalView)
		}
		return modalView
	}

	var b strings.Builder

	// 1. Header
	logo := LogoStyle.Render("cmdpp")
	title := TitleStyle.Render("Command Launcher")
	count := CountBadgeStyle.Render(fmt.Sprintf("(%d/%d commands)", len(m.filtered), len(m.commands)))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, logo, title, count) + "\n\n")

	// 2. Search Input
	searchStyle := SearchBoxStyle
	if m.searchFocused {
		searchStyle = SearchBoxFocusedStyle
	}
	searchBox := searchStyle.Render("🔍 " + m.searchInput.View())
	b.WriteString(searchBox + "\n")

	// 3. Command List
	if len(m.filtered) == 0 {
		if len(m.commands) == 0 {
			b.WriteString(NormalItemStyle.Render(
				"No commands stored yet.\n\nPress " + HelpKeyStyle.Render("[a]") + " to add your first command!\n",
			))
		} else {
			b.WriteString(NormalItemStyle.Render("No matching commands found. Press " + HelpKeyStyle.Render("[Esc]") + " to clear search.\n"))
		}
	} else {
		visibleEnd := min(len(m.filtered), m.viewportTop+m.maxVisibleRows)
		for i := m.viewportTop; i < visibleEnd; i++ {
			c := m.filtered[i]
			isSelected := (i == m.cursor)

			var cursorPrefix string
			if isSelected {
				cursorPrefix = CursorStyle.Render("▸ ")
			} else {
				cursorPrefix = "  "
			}

			// Format row
			nameStr := c.Name
			if isSelected {
				nameStr = ItemNameSelected.Render(fmt.Sprintf("%-18s", nameStr))
			} else {
				nameStr = ItemNameNormal.Render(fmt.Sprintf("%-18s", nameStr))
			}

			cmdPreview := c.Cmd
			maxCmdLen := max(20, m.width-45)
			if len(cmdPreview) > maxCmdLen {
				cmdPreview = cmdPreview[:maxCmdLen-3] + "..."
			}

			if isSelected {
				cmdPreview = ItemCmdSelected.Render(cmdPreview)
			} else {
				cmdPreview = ItemCmdNormal.Render(cmdPreview)
			}

			lastRun := cli.FormatRelativeTime(c.LastRunAt)
			meta := ItemMetaStyle.Render(fmt.Sprintf("%dx • %s", c.RunCount, lastRun))

			line := fmt.Sprintf("%s%s %s", cursorPrefix, nameStr, cmdPreview)
			if m.width > 70 {
				line = fmt.Sprintf("%-60s %s", line, meta)
			}

			if isSelected {
				b.WriteString(SelectedItemStyle.Render(line) + "\n")
			} else {
				b.WriteString(NormalItemStyle.Render(line) + "\n")
			}
		}

		if len(m.filtered) > m.maxVisibleRows {
			scrollInfo := fmt.Sprintf(" Showing %d-%d of %d ", m.viewportTop+1, visibleEnd, len(m.filtered))
			b.WriteString(ItemMetaStyle.Render(scrollInfo) + "\n")
		}
	}

	// 4. Details Box (for selected command)
	if len(m.filtered) > 0 && m.cursor < len(m.filtered) {
		selected := m.filtered[m.cursor]
		var detailContent strings.Builder
		detailContent.WriteString(DetailLabelStyle.Render("Command: ") + DetailCmdStyle.Render(selected.Cmd) + "\n")
		if selected.Description != "" {
			detailContent.WriteString(DetailLabelStyle.Render("About:   ") + DetailValueStyle.Render(selected.Description) + "\n")
		}
		lastRun := "Never"
		if selected.LastRunAt != nil {
			lastRun = fmt.Sprintf("%s (%s)", selected.LastRunAt.Format("2006-01-02 15:04"), cli.FormatRelativeTime(selected.LastRunAt))
		}
		stats := fmt.Sprintf("%s Created: %s  |  Last Run: %s  |  Executed: %d times",
			DetailLabelStyle.Render("Stats:  "),
			selected.CreatedAt.Format("2006-01-02"),
			lastRun,
			selected.RunCount,
		)
		detailContent.WriteString(DetailValueStyle.Render(stats))

		paneWidth := max(40, m.width-4)
		detailPane := DetailPaneStyle.Width(paneWidth).Render(detailContent.String())
		b.WriteString("\n" + detailPane + "\n")
	}

	// 5. Toast / Notification / Confirm Delete
	if m.isConfirmingDel && len(m.filtered) > 0 {
		sel := m.filtered[m.cursor]
		prompt := fmt.Sprintf("⚠️  Are you sure you want to delete %q?  [y] Yes   [n/Esc] Cancel", sel.Name)
		b.WriteString("\n" + ToastDangerStyle.Render(prompt) + "\n")
	} else if m.toastMsg != "" {
		if m.toastIsErr {
			b.WriteString("\n" + ToastDangerStyle.Render("⚠️  "+m.toastMsg) + "\n")
		} else {
			b.WriteString("\n" + ToastSuccessStyle.Render(m.toastMsg) + "\n")
		}
	}

	// 6. Footer / Keybindings
	var helpBar string
	if m.searchFocused {
		helpBar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			HelpKeyStyle.Render("[Enter]"), HelpDescStyle.Render(" Run/Accept   "),
			HelpKeyStyle.Render("[↑/↓]"), HelpDescStyle.Render(" Navigate   "),
			HelpKeyStyle.Render("[Esc]"), HelpDescStyle.Render(" Exit Search"),
		)
	} else {
		helpBar = lipgloss.JoinHorizontal(
			lipgloss.Left,
			HelpKeyStyle.Render("[Enter]"), HelpDescStyle.Render(" Run   "),
			HelpKeyStyle.Render("[/]"), HelpDescStyle.Render(" Search   "),
			HelpKeyStyle.Render("[a]"), HelpDescStyle.Render(" Add   "),
			HelpKeyStyle.Render("[e]"), HelpDescStyle.Render(" Edit   "),
			HelpKeyStyle.Render("[d]"), HelpDescStyle.Render(" Delete   "),
			HelpKeyStyle.Render("[c]"), HelpDescStyle.Render(" Copy   "),
			HelpKeyStyle.Render("[q]"), HelpDescStyle.Render(" Quit"),
		)
	}
	b.WriteString("\n" + HelpBarStyle.Render(helpBar))

	return b.String()
}

// Action returns the action resulting from the TUI run.
func (m Model) Action() AppAction {
	return m.action
}

// ChosenCommand returns the command selected for execution.
func (m Model) ChosenCommand() *store.Command {
	return m.chosenCommand
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
