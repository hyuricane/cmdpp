package cli

import (
	"bufio"
	"fmt"
	"io"
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

// ParseCLIArgs separates parameter overrides (KEY=VALUE) from trailing command arguments.
func ParseCLIArgs(params []store.Param, rawArgs []string) (map[string]string, []string) {
	paramNames := make(map[string]bool, len(params))
	for _, p := range params {
		paramNames[p.Name] = true
	}

	overrides := make(map[string]string)
	var extraArgs []string

	for _, arg := range rawArgs {
		if eqIdx := strings.IndexByte(arg, '='); eqIdx != -1 {
			k := arg[:eqIdx]
			v := arg[eqIdx+1:]
			if paramNames[k] {
				overrides[k] = v
				continue
			}
		}
		extraArgs = append(extraArgs, arg)
	}

	return overrides, extraArgs
}

// PromptParam interactively prompts the user on the terminal for a parameter value.
func PromptParam(in io.Reader, out io.Writer, p store.Param) (string, error) {
	reader := bufio.NewReader(in)
	for {
		if p.HasDefault {
			fmt.Fprintf(out, "%s [%s]: ", cyanStyle.Render(p.Name), dimStyle.Render(p.DefaultValue))
		} else {
			fmt.Fprintf(out, "%s: ", cyanStyle.Render(p.Name))
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		line = strings.TrimRight(line, "\r\n")
		line = strings.TrimSpace(line)

		if line == "" {
			if p.HasDefault {
				return p.DefaultValue, nil
			}
			fmt.Fprintf(out, "%s\n", errStyle.Render("Parameter is required."))
			continue
		}
		return line, nil
	}
}

// RunDirect executes the given command directly with optional extra arguments.
func RunDirect(s *store.Store, cmd *store.Command, extraArgs []string) (int, error) {
	// Append any additional arguments provided on the command line
	fullCmd := cmd.Cmd
	if len(extraArgs) > 0 {
		fullCmd = fmt.Sprintf("%s %s", fullCmd, strings.Join(extraArgs, " "))
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

// Run executes the stored command by name, resolving parameters from CLI args or prompting.
func Run(s *store.Store, args []string) (int, error) {
	if len(args) == 0 {
		return 1, fmt.Errorf("missing command name to run\nUsage: cmdpp run <command-name> [KEY=VALUE...] [additional args...]")
	}

	name := strings.TrimSpace(args[0])
	cmd, ok := s.Get(name)
	if !ok {
		return 1, fmt.Errorf("command %q not found. Run 'cmdpp list' to see all saved commands, or 'cmdpp' for interactive picker", name)
	}

	params := store.ExtractParams(cmd.Cmd)
	if len(params) == 0 {
		return RunDirect(s, cmd, args[1:])
	}

	overrides, extraArgs := ParseCLIArgs(params, args[1:])

	// Prompt for missing parameters if running in interactive terminal
	isTerm := IsTerminal(os.Stdin.Fd())
	for _, p := range params {
		if _, exists := overrides[p.Name]; exists {
			continue
		}

		if isTerm {
			val, err := PromptParam(os.Stdin, os.Stderr, p)
			if err != nil {
				if err == io.EOF {
					fmt.Fprintln(os.Stderr)
				}
				return 130, nil // User canceled
			}
			overrides[p.Name] = val
		} else {
			if p.HasDefault {
				overrides[p.Name] = p.DefaultValue
			} else {
				return 1, fmt.Errorf("missing required parameter %q; provide it via %s=<value>", p.Name, p.Name)
			}
		}
	}

	substitutedCmd := store.SubstituteParams(cmd.Cmd, overrides)
	cmdCopy := *cmd
	cmdCopy.Cmd = substitutedCmd

	return RunDirect(s, &cmdCopy, extraArgs)
}
