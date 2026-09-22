package cli

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#BD93F9"))
	subStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8F8F2"))
	cmdHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#50FA7B"))
	flagStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))
	exStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
)

// PrintHelp prints the usage manual for cmdpp.
func PrintHelp(out io.Writer) {
	fmt.Fprintf(out, `
%s - %s

%s
  cmdpp [command] [options]
  cmdpp                   Launch interactive TUI picker

%s
  %s  Add or update a command
  %s  Execute a stored command
  %s   Remove a stored command (aliases: rm, delete)
  %s  List all stored commands (uses 'more' if available)
  %s  Show this help screen
  %s  Show cmdpp version

%s
  %s
    Add a command to store:
    $ %s

  %s
    Add with description:
    $ %s

  %s
    Execute stored command:
    $ %s

  %s
    Execute with extra flags:
    $ %s

  %s
    List stored commands (pages with 'more'):
    $ %s

  %s
    Delete a stored command:
    $ %s

  %s
    Launch interactive TUI app to search & pick a command:
    $ %s

%s
  Inside the TUI picker:
  • %s / %s    Navigate commands
  • %s         Execute selected command
  • %s         Search & filter commands
  • %s         Add a new command
  • %s         Edit selected command
  • %s         Delete selected command
  • %s         Copy command to clipboard
  • %s         Quit TUI

`,
		titleStyle.Render("cmdpp"),
		subStyle.Render("Command Plus Plus - Store, manage, and trigger CLI commands with ease"),
		cmdHeader.Render("USAGE:"),
		cmdHeader.Render("COMMANDS:"),
		exStyle.Render("add <name> --cmd \"<cmd>\""),
		exStyle.Render("run <name> [args...]    "),
		exStyle.Render("rm <name>               "),
		exStyle.Render("list [--no-pager]       "),
		exStyle.Render("help, --help, -h        "),
		exStyle.Render("version, -v             "),
		cmdHeader.Render("EXAMPLES:"),
		flagStyle.Render("1. Add a command:"),
		subStyle.Render(`cmdpp add tunnel-one --cmd "ssh -N -L user@ssh.example.org"`),
		flagStyle.Render("2. Add with description:"),
		subStyle.Render(`cmdpp add tunnel-one --cmd "ssh -N -L user@ssh.example.org" --desc "Internal jump tunnel"`),
		flagStyle.Render("3. Run a command:"),
		subStyle.Render(`cmdpp run tunnel-one`),
		flagStyle.Render("4. Run with extra arguments:"),
		subStyle.Render(`cmdpp run tunnel-one -v`),
		flagStyle.Render("5. List commands:"),
		subStyle.Render(`cmdpp list`),
		flagStyle.Render("6. Delete a command:"),
		subStyle.Render(`cmdpp rm tunnel-one`),
		flagStyle.Render("7. Interactive TUI:"),
		subStyle.Render(`cmdpp`),
		cmdHeader.Render("TUI SHORTCUTS:"),
		exStyle.Render("↑/↓"), exStyle.Render("j/k"),
		exStyle.Render("Enter"),
		exStyle.Render("/"),
		exStyle.Render("a"),
		exStyle.Render("e"),
		exStyle.Render("d"),
		exStyle.Render("c"),
		exStyle.Render("q / Esc"),
	)
}
