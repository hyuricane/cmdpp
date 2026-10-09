package store

import (
	"reflect"
	"testing"
)

func TestExtractParams(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected []Param
	}{
		{
			name:     "no params",
			cmd:      "git status",
			expected: nil,
		},
		{
			name: "single param with default",
			cmd:  "ssh -N -L ${PORT:-3000}:127.0.0.1:${PORT:-3000} user@host.com",
			expected: []Param{
				{Name: "PORT", DefaultValue: "3000", HasDefault: true},
			},
		},
		{
			name: "multiple unique params",
			cmd:  "docker run -p ${HOST_PORT:-8080}:${CONTAINER_PORT:-80} -e ENV=${ENV} myimage",
			expected: []Param{
				{Name: "HOST_PORT", DefaultValue: "8080", HasDefault: true},
				{Name: "CONTAINER_PORT", DefaultValue: "80", HasDefault: true},
				{Name: "ENV", DefaultValue: "", HasDefault: false},
			},
		},
		{
			name: "escaped param syntax",
			cmd:  `echo "\${ESCAPED:-foo}" and ${REAL:-bar}`,
			expected: []Param{
				{Name: "REAL", DefaultValue: "bar", HasDefault: true},
			},
		},
		{
			name: "empty default",
			cmd:  "cmd --opt=${OPT:-}",
			expected: []Param{
				{Name: "OPT", DefaultValue: "", HasDefault: true},
			},
		},
		{
			name: "default containing colons and slashes",
			cmd:  "curl ${URL:-http://localhost:8080/api/v1} -H 'Auth: ${TOKEN}'",
			expected: []Param{
				{Name: "URL", DefaultValue: "http://localhost:8080/api/v1", HasDefault: true},
				{Name: "TOKEN", DefaultValue: "", HasDefault: false},
			},
		},
		{
			name: "multiple occurrences where later occurrence adds default",
			cmd:  "cmd ${PARAM} --fallback ${PARAM:-default_val}",
			expected: []Param{
				{Name: "PARAM", DefaultValue: "default_val", HasDefault: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractParams(tt.cmd)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ExtractParams(%q) = %+v, want %+v", tt.cmd, got, tt.expected)
			}
		})
	}
}

func TestSubstituteParams(t *testing.T) {
	cmd := "ssh -N -L ${PORT:-3000}:127.0.0.1:${PORT:-3000} ${USER:-admin}@${HOST}"

	tests := []struct {
		name     string
		values   map[string]string
		expected string
	}{
		{
			name:     "use all defaults and missing required empty",
			values:   nil,
			expected: "ssh -N -L 3000:127.0.0.1:3000 admin@",
		},
		{
			name: "override default and provide required",
			values: map[string]string{
				"PORT": "8080",
				"HOST": "remote.server.org",
			},
			expected: "ssh -N -L 8080:127.0.0.1:8080 admin@remote.server.org",
		},
		{
			name: "override everything",
			values: map[string]string{
				"PORT": "9000",
				"USER": "deploy",
				"HOST": "prod.server.org",
			},
			expected: "ssh -N -L 9000:127.0.0.1:9000 deploy@prod.server.org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SubstituteParams(cmd, tt.values)
			if got != tt.expected {
				t.Errorf("SubstituteParams() = %q, want %q", got, tt.expected)
			}
		})
	}

	// Test unescaping
	escapedCmd := `echo "\${LITERAL}" ${PARAM:-val}`
	gotEscaped := SubstituteParams(escapedCmd, map[string]string{"PARAM": "real"})
	expectedEscaped := `echo "${LITERAL}" real`
	if gotEscaped != expectedEscaped {
		t.Errorf("SubstituteParams() unescaping = %q, want %q", gotEscaped, expectedEscaped)
	}
}
