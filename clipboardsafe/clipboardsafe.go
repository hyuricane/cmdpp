// Package clipboardsafe
// handle reading from clipboard safely
package clipboardsafe

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"time"

	"github.com/atotto/clipboard"
)

const (
	MaxClipboardBytes = 4 << 20 // 4 MiB
	MaxClipboardStr   = "4 MiB"
	ClipboardTimeout  = 1 * time.Second
)

// ReadAllLimited reads clipboard content with a hard byte ceiling and timeout.
func ReadAllLimited() (string, error) {
	// On Windows, atotto/clipboard uses direct Win32 API calls (user32.dll),
	// which is non-streaming and already capped to a fixed memory buffer.
	if runtime.GOOS == "windows" {
		return limitedAtotto()
	}

	if clipboard.Unsupported {
		return "", fmt.Errorf("clipboard unsupported")
	}

	ctx, cancel := context.WithTimeout(context.Background(), ClipboardTimeout)
	defer cancel()

	var cmd *exec.Cmd
	switch {
	case runtime.GOOS == "darwin" && lookpath("pbpaste"): // macOS
		cmd = exec.CommandContext(ctx, "pbpaste")
	case lookpath("wl-paste"): // Wayland
		cmd = exec.CommandContext(ctx, "wl-paste", "--no-newline")
	case lookpath("xclip"): // X11 xclip
		cmd = exec.CommandContext(ctx, "xclip", "-out", "-selection", "clipboard")
	case lookpath("xsel"): // X11 xsel
		cmd = exec.CommandContext(ctx, "xsel", "--output", "--clipboard")
	default:
		return limitedAtotto()
	}

	return readFromCmd(ctx, cmd)
}

func readFromCmd(ctx context.Context, cmd *exec.Cmd) (string, error) {
	setProcessGroup(cmd)

	out, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = out.Close()
		case <-done:
		}
	}()

	limited := io.LimitReader(out, MaxClipboardBytes+1)
	data, err := io.ReadAll(limited)

	// cleanup spawned processes first
	_ = out.Close()
	killProcessGroup(cmd)
	_ = cmd.Wait()

	if ctx.Err() != nil {
		return "", fmt.Errorf("clipboard read timed out")
	}

	if err != nil {
		return "", err
	}
	if int64(len(data)) > MaxClipboardBytes {
		return "", fmt.Errorf("clipboard data exceeds %s limit", MaxClipboardStr)
	}

	return string(data), nil
}

func limitedAtotto() (string, error) {
	s, err := clipboard.ReadAll()
	if err != nil {
		return "", err
	}
	if len(s) > MaxClipboardBytes {
		return "", fmt.Errorf("clipboard data exceeds %s limit", MaxClipboardStr)
	}
	return s, nil
}

func lookpath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}
