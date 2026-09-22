package tui

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// CopyToClipboard copies the given text to the system clipboard using wl-copy, xclip, xsel, or OSC 52.
func CopyToClipboard(text string) error {
	// Emit OSC 52 sequence directly to terminal (supported by modern terminal emulators)
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	osc52 := fmt.Sprintf("\033]52;c;%s\a", encoded)
	_, _ = os.Stdout.WriteString(osc52)

	// Try wl-copy (Wayland)
	if wlCopy, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command(wlCopy)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try xclip (X11)
	if xclip, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(xclip, "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	// Try xsel (X11)
	if xsel, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command(xsel, "--clipboard", "--input")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	return nil
}
