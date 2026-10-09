package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hyuricane/cmdpp/store"
)

func TestParseCLIArgs(t *testing.T) {
	params := []store.Param{
		{Name: "PORT", DefaultValue: "3000", HasDefault: true},
		{Name: "HOST", DefaultValue: "", HasDefault: false},
	}

	rawArgs := []string{"PORT=8080", "--verbose", "HOST=example.org", "extra-flag"}
	overrides, extra := ParseCLIArgs(params, rawArgs)

	if overrides["PORT"] != "8080" {
		t.Errorf("expected PORT=8080, got %q", overrides["PORT"])
	}
	if overrides["HOST"] != "example.org" {
		t.Errorf("expected HOST=example.org, got %q", overrides["HOST"])
	}
	if len(extra) != 2 || extra[0] != "--verbose" || extra[1] != "extra-flag" {
		t.Errorf("unexpected extra args: %+v", extra)
	}
}

func TestPromptParamWithDefault(t *testing.T) {
	param := store.Param{
		Name:         "PORT",
		DefaultValue: "3000",
		HasDefault:   true,
	}

	// Case 1: user presses Enter (empty input) -> uses default
	input := strings.NewReader("\n")
	output := &bytes.Buffer{}
	val, err := PromptParam(input, output, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "3000" {
		t.Errorf("expected default value '3000', got %q", val)
	}

	// Case 2: user types a value
	input = strings.NewReader("9000\n")
	output = &bytes.Buffer{}
	val, err = PromptParam(input, output, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "9000" {
		t.Errorf("expected value '9000', got %q", val)
	}
}

func TestPromptParamRequired(t *testing.T) {
	param := store.Param{
		Name:         "TOKEN",
		DefaultValue: "",
		HasDefault:   false,
	}

	// User types empty first, then types "secret"
	input := strings.NewReader("\nsecret\n")
	output := &bytes.Buffer{}
	val, err := PromptParam(input, output, param)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret" {
		t.Errorf("expected 'secret', got %q", val)
	}
}
