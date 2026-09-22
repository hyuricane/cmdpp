package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"cmdpp/store"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	borderStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	runBadge     = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	statStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
	descStyle    = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#F1FA8C"))
	cmdLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
)

// FormatRelativeTime returns a human-readable representation of a timestamp.
func FormatRelativeTime(t *time.Time) string {
	if t == nil {
		return "never"
	}
	diff := time.Since(*t)
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	}
	days := int(diff.Hours() / 24)
	if days == 1 {
		return "yesterday"
	}
	if days < 7 {
		return fmt.Sprintf("%dd ago", days)
	}
	return t.Format("2006-01-02")
}

// List handles the 'cmdpp list' subcommand.
func List(s *store.Store, out io.Writer, args []string) error {
	noPager := false
	sortOrder := "name"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--no-pager" || arg == "-n" {
			noPager = true
		} else if arg == "--sort" || arg == "-s" {
			if i+1 < len(args) {
				sortOrder = args[i+1]
				i++
			}
		} else if strings.HasPrefix(arg, "--sort=") {
			sortOrder = strings.TrimPrefix(arg, "--sort=")
		}
	}

	commands := s.SortedList(sortOrder)
	if len(commands) == 0 {
		var b bytes.Buffer
		fmt.Fprintf(&b, "\n%s\n", dimStyle.Render("No commands stored yet."))
		fmt.Fprintf(&b, "Add your first command with:\n")
		fmt.Fprintf(&b, "  %s\n\n", cyanStyle.Render("cmdpp add <name> --cmd \"<command>\" [--desc \"<description>\"]"))
		return DisplayWithPager(out, b.String(), noPager)
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "\n%s (%d total):\n\n", headerStyle.Render("Stored Commands"), len(commands))

	for i, c := range commands {
		lastRun := FormatRelativeTime(c.LastRunAt)
		runInfo := fmt.Sprintf("runs: %d • last: %s", c.RunCount, lastRun)

		fmt.Fprintf(&buf, "  %s  %s\n",
			cyanStyle.Render(fmt.Sprintf("%-20s", c.Name)),
			statStyle.Render(runInfo),
		)
		fmt.Fprintf(&buf, "    %s %s\n",
			dimStyle.Render("$"),
			cmdLineStyle.Render(c.Cmd),
		)
		if c.Description != "" {
			fmt.Fprintf(&buf, "    %s\n", descStyle.Render("• "+c.Description))
		}

		if i < len(commands)-1 {
			fmt.Fprintf(&buf, "    %s\n", borderStyle.Render(strings.Repeat("─", 50)))
		}
	}
	fmt.Fprintf(&buf, "\n%s\n\n", dimStyle.Render("Run any command with: cmdpp run <name>   |   Launch interactive picker: cmdpp"))

	return DisplayWithPager(out, buf.String(), noPager)
}
