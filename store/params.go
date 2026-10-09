package store

import (
	"strings"
)

// Param represents a named command parameter extracted from a command template.
type Param struct {
	Name         string // Name of the parameter, e.g. "PORT"
	DefaultValue string // Default value if specified, e.g. "3000"
	HasDefault   bool   // True if a default value syntax (":-") was used
}

// isValidIdent checks if the string is a valid parameter identifier [a-zA-Z_][a-zA-Z0-9_]*.
func isValidIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != '_' {
				return false
			}
		} else {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' {
				return false
			}
		}
	}
	return true
}

// parseParamContent parses the inside of ${...}.
// Examples:
//
//	"PORT:-3000" -> Param{Name: "PORT", DefaultValue: "3000", HasDefault: true}, true
//	"PORT"       -> Param{Name: "PORT", DefaultValue: "", HasDefault: false}, true
//	"PORT:-"     -> Param{Name: "PORT", DefaultValue: "", HasDefault: true}, true
func parseParamContent(content string) (Param, bool) {
	if sepIdx := strings.Index(content, ":-"); sepIdx != -1 {
		name := content[:sepIdx]
		defVal := content[sepIdx+2:]
		if isValidIdent(name) {
			return Param{
				Name:         name,
				DefaultValue: defVal,
				HasDefault:   true,
			}, true
		}
	} else {
		if isValidIdent(content) {
			return Param{
				Name:         content,
				DefaultValue: "",
				HasDefault:   false,
			}, true
		}
	}
	return Param{}, false
}

// ExtractParams parses a command string and returns unique named parameters in appearance order.
// Escaped syntax "\${VAR}" is ignored and not treated as a parameter.
func ExtractParams(cmd string) []Param {
	var params []Param
	seen := make(map[string]int) // name -> index in params

	i := 0
	n := len(cmd)
	for i < n {
		if i+1 < n && cmd[i] == '\\' && cmd[i+1] == '$' {
			// Escaped \$, skip both
			i += 2
			continue
		}
		if i+1 < n && cmd[i] == '$' && cmd[i+1] == '{' {
			closeIdx := strings.IndexByte(cmd[i+2:], '}')
			if closeIdx == -1 {
				i += 2
				continue
			}
			content := cmd[i+2 : i+2+closeIdx]
			param, ok := parseParamContent(content)
			if ok {
				if idx, exists := seen[param.Name]; exists {
					// If previously seen without default, but this occurrence has default, update it
					if !params[idx].HasDefault && param.HasDefault {
						params[idx].HasDefault = true
						params[idx].DefaultValue = param.DefaultValue
					}
				} else {
					seen[param.Name] = len(params)
					params = append(params, param)
				}
			}
			i += 2 + closeIdx + 1
			continue
		}
		i++
	}
	return params
}

// HasParams returns true if the command template contains at least one parameter.
func HasParams(cmd string) bool {
	return len(ExtractParams(cmd)) > 0
}

// SubstituteParams replaces all parameter placeholders in cmd with resolved values.
// If a parameter is not in values or its value is empty, its default value is used if available.
// Escaped "\${" is unescaped to "${".
func SubstituteParams(cmd string, values map[string]string) string {
	var sb strings.Builder
	i := 0
	n := len(cmd)
	for i < n {
		if i+1 < n && cmd[i] == '\\' && cmd[i+1] == '$' {
			// Unescape \${ to ${
			sb.WriteByte('$')
			i += 2
			continue
		}
		if i+1 < n && cmd[i] == '$' && cmd[i+1] == '{' {
			closeIdx := strings.IndexByte(cmd[i+2:], '}')
			if closeIdx != -1 {
				content := cmd[i+2 : i+2+closeIdx]
				param, ok := parseParamContent(content)
				if ok {
					val, hasVal := values[param.Name]
					if hasVal && val != "" {
						sb.WriteString(val)
					} else if param.HasDefault {
						sb.WriteString(param.DefaultValue)
					} else if hasVal {
						// Explicitly empty
						sb.WriteString(val)
					}
					i += 2 + closeIdx + 1
					continue
				}
			}
		}
		sb.WriteByte(cmd[i])
		i++
	}
	return sb.String()
}
