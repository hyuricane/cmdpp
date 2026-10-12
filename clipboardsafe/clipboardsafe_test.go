package clipboardsafe

import (
	"context"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestReadFromCmd_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix command test skipped on Windows")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "echo", "-n", "hello clipboard")
	out, err := readFromCmd(ctx, cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello clipboard" {
		t.Fatalf("expected 'hello clipboard', got %q", out)
	}
}

func TestReadFromCmd_TimeoutWithOrphanedChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Process group / fork test is Unix-specific")
	}

	// This simulates wl-paste forking a child process (like cat) that holds
	// the stdout pipe open and blocks for 5 seconds.
	// Without process group killing and closing the pipe, this would block
	// for the full 5 seconds even after the parent is killed.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", "sleep 5 & wait")
	start := time.Now()
	_, err := readFromCmd(ctx, cmd)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected 'timed out' error, got: %v", err)
	}
	if elapsed > 1*time.Second {
		t.Fatalf("test took %v; expected to abort within ~100ms deadline despite child holding pipe", elapsed)
	}
}

func TestReadFromCmd_ExceedsLimit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix command test skipped on Windows")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Generate 5 MiB of data (exceeding MaxClipboardBytes = 4 MiB)
	cmd := exec.CommandContext(ctx, "sh", "-c", "head -c 5242880 /dev/zero")
	_, err := readFromCmd(ctx, cmd)
	if err == nil {
		t.Fatal("expected limit exceeded error, got nil")
	}
	if !strings.Contains(err.Error(), "limit") {
		t.Fatalf("expected limit exceeded error, got: %v", err)
	}
}
