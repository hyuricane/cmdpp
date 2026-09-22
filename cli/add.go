package cli

import (
	"fmt"
	"io"
	"strings"

	"cmdpp/store"
	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#04B575"))
	cyanStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00D7D7"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#8A8A8A"))
	errStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5F87"))
)

// Add handles the 'cmdpp add' subcommand.
// Supports:
//   cmdpp add <name> --cmd "<command>" [--desc "<description>"]
//   cmdpp add <name> -c "<command>" [-d "<description>"]
//   cmdpp add --cmd "<command>" <name>
//   cmdpp add <name> "<command>"
func Add(s *store.Store, out io.Writer, args []string) error {
	var (
		name        string
		cmdStr      string
		desc        string
		positionals []string
	)

	i := 0
	for i < len(args) {
		arg := args[i]
		if arg == "--cmd" || arg == "-c" {
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires an argument", arg)
			}
			cmdStr = args[i+1]
			i += 2
			continue
		} else if strings.HasPrefix(arg, "--cmd=") {
			cmdStr = strings.TrimPrefix(arg, "--cmd=")
			i++
			continue
		} else if strings.HasPrefix(arg, "-c=") {
			cmdStr = strings.TrimPrefix(arg, "-c=")
			i++
			continue
		} else if arg == "--desc" || arg == "--description" || arg == "-d" {
			if i+1 >= len(args) {
				return fmt.Errorf("flag %s requires an argument", arg)
			}
			desc = args[i+1]
			i += 2
			continue
		} else if strings.HasPrefix(arg, "--desc=") {
			desc = strings.TrimPrefix(arg, "--desc=")
			i++
			continue
		} else if strings.HasPrefix(arg, "--description=") {
			desc = strings.TrimPrefix(arg, "--description=")
			i++
			continue
		} else if strings.HasPrefix(arg, "-d=") {
			desc = strings.TrimPrefix(arg, "-d=")
			i++
			continue
		} else if strings.HasPrefix(arg, "-") {
			return fmt.Errorf("unknown flag: %s\nUsage: cmdpp add <name> --cmd \"<command>\" [--desc \"<description>\"]", arg)
		} else {
			positionals = append(positionals, arg)
			i++
		}
	}

	if len(positionals) == 0 {
		return fmt.Errorf("missing command name\nUsage: cmdpp add <name> --cmd \"<command>\" [--desc \"<description>\"]")
	}

	name = strings.TrimSpace(positionals[0])

	// If --cmd flag was not provided, check if a positional command string was passed
	if cmdStr == "" && len(positionals) > 1 {
		cmdStr = strings.TrimSpace(strings.Join(positionals[1:], " "))
	}

	if cmdStr == "" {
		return fmt.Errorf("command string cannot be empty\nUsage: cmdpp add %s --cmd \"<command>\"", name)
	}

	cmd, updated, err := s.AddOrUpdate(name, cmdStr, desc)
	if err != nil {
		return err
	}

	if err := s.Save(); err != nil {
		return fmt.Errorf("failed to save command: %w", err)
	}

	if updated {
		fmt.Fprintf(out, "%s Updated command %s\n", successStyle.Render("✓"), cyanStyle.Render(cmd.Name))
	} else {
		fmt.Fprintf(out, "%s Saved command %s\n", successStyle.Render("✓"), cyanStyle.Render(cmd.Name))
	}

	fmt.Fprintf(out, "  %s %s\n", dimStyle.Render("Command:"), cmd.Cmd)
	if cmd.Description != "" {
		fmt.Fprintf(out, "  %s %s\n", dimStyle.Render("Description:"), cmd.Description)
	}
	fmt.Fprintf(out, "  %s %s\n\n", dimStyle.Render("Run it anytime with:"), cyanStyle.Render("cmdpp run "+cmd.Name))

	return nil
}
