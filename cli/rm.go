package cli

import (
	"fmt"
	"io"
	"strings"

	"cmdpp/store"
)

// Rm handles the 'cmdpp rm' (and remove/delete) subcommand.
func Rm(s *store.Store, out io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command name to remove\nUsage: cmdpp rm <command-name>")
	}

	name := strings.TrimSpace(args[0])
	removed, err := s.Remove(name)
	if err != nil {
		return err
	}

	if !removed {
		return fmt.Errorf("command %q not found. Run 'cmdpp list' to see available commands", name)
	}

	if err := s.Save(); err != nil {
		return fmt.Errorf("failed to save changes: %w", err)
	}

	fmt.Fprintf(out, "%s Deleted command %s\n", successStyle.Render("✓"), cyanStyle.Render(name))
	return nil
}
