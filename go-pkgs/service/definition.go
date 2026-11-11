package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Definition is one managed command.
//
// Detached marks a process the supervisor does not own the lifetime of: it is
// started in its own session so it keeps running after the host exits, and
// Shutdown leaves it alone. Hosts use it for services the user asked to keep
// serving independently (eluc's standalone llm-proxy); an attached service
// dies with its manager, which is what ai-critic's child services do.
type Definition struct {
	ID         string            `json:"id"`
	Name       string            `json:"name,omitempty"`
	Command    string            `json:"command"`
	WorkingDir string            `json:"workingDir,omitempty"`
	ExtraEnv   map[string]string `json:"extraEnv,omitempty"`
	ExtraPath  string            `json:"extraPath,omitempty"`
	// InheritInteractiveShellEnv merges the environment of an interactive
	// login shell, which is what sources ~/.zshrc and ~/.bashrc.
	InheritInteractiveShellEnv bool  `json:"inheritInteractiveShellEnv,omitempty"`
	Enabled                    *bool `json:"enabled,omitempty"`
	Detached                   bool  `json:"detached,omitempty"`
}

// Bool returns a pointer to v, for Definition.Enabled.
func Bool(v bool) *bool { return &v }

// IsEnabled reports whether the definition is allowed to run. An absent
// Enabled means enabled: definitions written before the switch existed must
// keep starting.
func (d Definition) IsEnabled() bool {
	return d.Enabled == nil || *d.Enabled
}

// NormalizeDefinition trims user input and resolves the parts that have a
// canonical form: Name defaults to ID, WorkingDir to the user's home, and
// ExtraPath is cleaned exactly like a PATH entry.
//
// An empty ID is allowed here (hosts may generate one from the name); use
// ValidateDefinition before persisting or starting.
func NormalizeDefinition(d Definition) (Definition, error) {
	d.ID = strings.TrimSpace(d.ID)
	d.Name = strings.TrimSpace(d.Name)
	if d.Name == "" {
		d.Name = d.ID
	}
	d.Command = strings.TrimSpace(d.Command)
	d.ExtraPath = strings.TrimSpace(d.ExtraPath)
	d.ExtraEnv = NormalizeEnv(d.ExtraEnv)

	extraPath, err := NormalizeExtraPath(d.ExtraPath)
	if err != nil {
		return d, fmt.Errorf("extra path: %w", err)
	}
	d.ExtraPath = extraPath
	d.WorkingDir = NormalizeWorkingDir(d.WorkingDir)
	return d, nil
}

// NormalizeWorkingDir resolves an empty or relative working directory to an
// absolute one. An empty value means the user's home directory; when home is
// unavailable it falls back to the process working directory.
func NormalizeWorkingDir(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			dir = home
		} else if cwd, err := os.Getwd(); err == nil {
			dir = cwd
		}
	}
	if dir == "" {
		return ""
	}
	if !filepath.IsAbs(dir) {
		if abs, err := filepath.Abs(dir); err == nil {
			dir = abs
		}
	}
	return dir
}

// EnsureWorkingDir creates dir (including parents) when non-empty.
func EnsureWorkingDir(dir string) error {
	if strings.TrimSpace(dir) == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create working directory %s: %w", dir, err)
	}
	return nil
}

// ValidateDefinition checks the fields a command needs to run.
func ValidateDefinition(d Definition) error {
	if strings.TrimSpace(d.ID) == "" {
		return fmt.Errorf("service id is required")
	}
	if strings.TrimSpace(d.Name) == "" {
		return fmt.Errorf("service name is required")
	}
	if strings.TrimSpace(d.Command) == "" {
		return fmt.Errorf("service command is required")
	}
	for key := range d.ExtraEnv {
		if !IsValidEnvKey(key) {
			return fmt.Errorf("invalid environment variable name: %s", key)
		}
	}
	if _, err := NormalizeExtraPath(d.ExtraPath); err != nil {
		return fmt.Errorf("extra path: %w", err)
	}
	return nil
}

// IsValidEnvKey reports whether key is a POSIX environment variable name.
func IsValidEnvKey(key string) bool {
	if key == "" {
		return false
	}
	for i, ch := range key {
		if i == 0 {
			if ch != '_' && !(ch >= 'A' && ch <= 'Z') && !(ch >= 'a' && ch <= 'z') {
				return false
			}
			continue
		}
		if ch != '_' && !(ch >= 'A' && ch <= 'Z') && !(ch >= 'a' && ch <= 'z') && !(ch >= '0' && ch <= '9') {
			return false
		}
	}
	return true
}

// NormalizeEnv trims keys and drops blank ones. A nil map stays nil so a
// definition round-trips without growing empty objects.
func NormalizeEnv(env map[string]string) map[string]string {
	if len(env) == 0 {
		return nil
	}
	normalized := make(map[string]string, len(env))
	for key, value := range env {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		normalized[key] = value
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

// CloneStringMap returns a copy of src, so a definition handed to the
// supervisor cannot be mutated through another reference.
func CloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	dst := make(map[string]string, len(src))
	for key, value := range src {
		dst[key] = value
	}
	return dst
}

// StringMapEqual compares two env maps.
func StringMapEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if b[key] != value {
			return false
		}
	}
	return true
}

// EnvValue returns the value of key in an os.Environ-style slice.
func EnvValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

// SetEnvValue sets key in an os.Environ-style slice, keeping the position of an
// existing entry and appending a new one otherwise.
func SetEnvValue(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	replaced := false
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			if replaced {
				continue
			}
			out = append(out, prefix+value)
			replaced = true
			continue
		}
		out = append(out, item)
	}
	if !replaced {
		out = append(out, prefix+value)
	}
	return out
}

// ShellQuote quotes s for bash -lc. Safe characters stay bare so commands keep
// their readable form in logs.
func ShellQuote(s string) string {
	safe := true
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '/' || c == '.' || c == '-' || c == '_') {
			safe = false
			break
		}
	}
	if safe {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// LogPath returns the log file a supervisor writes for id inside logDir.
func LogPath(logDir, id string) string {
	return filepath.Join(logDir, id+".log")
}

// DefinitionIDs returns the ids of defs in order.
func DefinitionIDs(defs []Definition) []string {
	ids := make([]string, 0, len(defs))
	for _, def := range defs {
		ids = append(ids, def.ID)
	}
	return ids
}

// SortDefinitions orders definitions by ID so listings are stable across
// config reloads.
func SortDefinitions(defs []Definition) {
	sort.Slice(defs, func(i, j int) bool { return defs[i].ID < defs[j].ID })
}
