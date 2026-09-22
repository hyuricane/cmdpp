package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hyuricane/cmdpp/store"
)

var (
	rocketStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6"))
	cmdBoxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
)

// Run executes the stored command by name.
func Run(s *store.Store, args []string) (int, error) {
	if len(args) == 0 {
		return 1, fmt.Errorf("missing command name to run\nUsage: cmdpp run <command-name> [additional args...]")
	}

	name := strings.TrimSpace(args[0])
	cmd, ok := s.Get(name)
	if !ok {
		return 1, fmt.Errorf("command %q not found. Run 'cmdpp list' to see all saved commands, or 'cmdpp' for interactive picker", name)
	}

	// Append any additional arguments provided on the command line
	fullCmd := cmd.Cmd
	if len(args) > 1 {
		fullCmd = fmt.Sprintf("%s %s", fullCmd, strings.Join(args[1:], " "))
	}

	// Update run stats in the background
	_ = s.RecordRun(cmd.Name)
	_ = s.Save()

	// Print execution banner
	fmt.Fprintf(os.Stderr, "%s %s: %s\n\n", rocketStyle.Render("🚀 Running"), cyanStyle.Render(cmd.Name), cmdBoxStyle.Render(fullCmd))

	// Determine shell to use
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	execCmd := exec.Command(shell, "-c", fullCmd)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("failed to run command: %w", err)
	}

	return 0, nil
}
