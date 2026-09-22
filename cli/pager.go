package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sys/unix"
)

// IsTerminal returns true if the given file descriptor is attached to a terminal.
func IsTerminal(fd uintptr) bool {
	_, err := unix.IoctlGetTermios(int(fd), unix.TCGETS)
	return err == nil
}

// DisplayWithPager outputs the given text using `more` if stdout is a TTY and `more` is available.
// Otherwise, it writes directly to w.
func DisplayWithPager(w io.Writer, text string, disablePager bool) error {
	if disablePager || !IsTerminal(os.Stdout.Fd()) {
		_, err := fmt.Fprint(w, text)
		return err
	}

	morePath, err := exec.LookPath("more")
	if err != nil {
		// "more" not found in PATH, fallback to direct stdout
		_, err := fmt.Fprint(w, text)
		return err
	}

	cmd := exec.Command(morePath)
	cmd.Stdin = strings.NewReader(text)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// If more fails or is interrupted by the user, fallback to stdout
		// Avoid erroring out if user simply pressed 'q'
		if exitErr, ok := err.(*exec.ExitError); ok {
			_ = exitErr
			return nil
		}
		_, _ = fmt.Fprint(w, text)
	}
	return nil
}
