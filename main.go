package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hyuricane/cmdpp/cli"
	"github.com/hyuricane/cmdpp/store"
	"github.com/hyuricane/cmdpp/tui"
)

var Version = "0.1.0-rc1"

func main() {
	s, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing cmdpp store: %v\n", err)
		os.Exit(1)
	}

	args := os.Args[1:]

	// If no arguments provided, or if explicitly invoked as 'cmdpp tui', launch TUI
	if len(args) == 0 || args[0] == "tui" {
		runTUI(s)
		return
	}

	cmd := args[0]
	cmdArgs := args[1:]

	switch cmd {
	case "add":
		if err := cli.Add(s, os.Stdout, cmdArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "run":
		exitCode, err := cli.Run(s, cmdArgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(exitCode)

	case "rm", "remove", "delete":
		if err := cli.Rm(s, os.Stdout, cmdArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "list", "ls":
		if err := cli.List(s, os.Stdout, cmdArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		cli.PrintHelp(os.Stdout)

	case "version", "--version", "-v":
		fmt.Printf("cmdpp v%s\n", strings.TrimPrefix(Version, "v"))

	default:
		// Check if the argument is actually the name of a stored command
		if _, exists := s.Get(cmd); exists {
			fmt.Fprintf(os.Stderr, "Hint: To execute %q, run:\n  cmdpp run %s\n\n", cmd, cmd)
		}
		fmt.Fprintf(os.Stderr, "Unknown command: %q\nRun 'cmdpp help' for usage instructions.\n", cmd)
		os.Exit(1)
	}
}

func runTUI(s *store.Store) {
	m := tui.NewModel(s)
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}

	if finalState, ok := finalModel.(tui.Model); ok {
		if finalState.Action() == tui.ActionRun && finalState.ChosenCommand() != nil {
			exitCode, err := cli.Run(s, []string{finalState.ChosenCommand().Name})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(exitCode)
		}
	}
}
