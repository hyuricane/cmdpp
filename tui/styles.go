package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Palette
	PrimaryColor   = lipgloss.Color("#BD93F9") // Purple / Dracula
	SecondaryColor = lipgloss.Color("#50FA7B") // Green
	AccentColor    = lipgloss.Color("#8BE9FD") // Cyan
	WarningColor   = lipgloss.Color("#FFB86C") // Orange
	DangerColor    = lipgloss.Color("#FF5555") // Red
	MutedColor     = lipgloss.Color("#6272A4") // Muted blue/gray
	TextColor      = lipgloss.Color("#F8F8F2") // Off-white
	BgDark         = lipgloss.Color("#21222C") // Dark background
	HighlightBg    = lipgloss.Color("#44475A") // Highlight background

	// Header styles
	LogoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#282A36")).
			Background(PrimaryColor).
			Padding(0, 1).
			MarginRight(1)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextColor)

	CountBadgeStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginLeft(1)

	// Search bar styles
	SearchBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(0, 1).
			MarginBottom(1)

	SearchBoxFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(AccentColor).
				Padding(0, 1).
				MarginBottom(1)

	CursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor)

	// List item styles
	NormalItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
				Background(HighlightBg).
				Bold(true).
				Padding(0, 1)

	ItemNameNormal = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextColor)

	ItemNameSelected = lipgloss.NewStyle().
				Bold(true).
				Foreground(AccentColor)

	ItemCmdNormal = lipgloss.NewStyle().
			Foreground(MutedColor)

	ItemCmdSelected = lipgloss.NewStyle().
			Foreground(SecondaryColor)

	ItemMetaStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Detail preview styles
	DetailPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(0, 1).
			MarginTop(1)

	DetailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(AccentColor)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(TextColor)

	DetailCmdStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	// Help / Footer styles
	HelpBarStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			MarginTop(1)

	HelpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(WarningColor)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(MutedColor)

	// Status / Toast styles
	ToastSuccessStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(SecondaryColor).
				MarginTop(1)

	ToastDangerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(DangerColor).
				MarginTop(1)

	// Modal styles
	ModalBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2).
			Width(64)

	ModalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	ModalFieldLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(AccentColor)

	ModalInputActive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(SecondaryColor).
				Padding(0, 1)

	ModalInputInactive = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(MutedColor).
				Padding(0, 1)

	// Dropdown suggestion styles
	DropdownContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(MutedColor).
				Padding(0, 1).
				MarginBottom(1)

	DropdownHeaderStyle = lipgloss.NewStyle().
				Foreground(AccentColor).
				Bold(true)

	DropdownItemNormal = lipgloss.NewStyle().
				Foreground(TextColor)

	DropdownItemSelected = lipgloss.NewStyle().
				Background(HighlightBg).
				Foreground(SecondaryColor).
				Bold(true)
)
