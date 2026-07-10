package wrkcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the optional user config at $WRK_HOME/config.json.
type Config struct {
	Version int            `json:"version"`
	Create  *CreateSection `json:"create,omitempty"`
}

// CreateSection holds create-mode options.
type CreateSection struct {
	Interceptor *CreateInterceptor `json:"interceptor,omitempty"`
}

// CreateInterceptor replaces native create with an external argv when Enabled.
type CreateInterceptor struct {
	Enabled bool                `json:"enabled"`
	Argv    []string            `json:"argv"`
	Vars    map[string]VarValue `json:"vars,omitempty"`
}

// VarValue is a JSON string or string array. Arrays expand element-wise then
// join with "\n" after template expansion.
type VarValue struct {
	IsList bool
	Scalar string
	Lines  []string
}

// UnmarshalJSON accepts a JSON string or array of strings.
func (v *VarValue) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return fmt.Errorf("empty var value")
	}
	if data[0] == '"' {
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		v.IsList = false
		v.Scalar = s
		v.Lines = nil
		return nil
	}
	if data[0] == '[' {
		var lines []string
		if err := json.Unmarshal(data, &lines); err != nil {
			return err
		}
		v.IsList = true
		v.Lines = lines
		v.Scalar = ""
		return nil
	}
	return fmt.Errorf("interceptor var must be string or array of strings")
}

// MarshalJSON emits a JSON string or array of strings.
func (v VarValue) MarshalJSON() ([]byte, error) {
	if v.IsList {
		return json.Marshal(v.Lines)
	}
	return json.Marshal(v.Scalar)
}

// loadConfig reads $WRK_HOME/config.json. Missing file returns (nil, nil).
func loadConfig(wrkHome string) (*Config, error) {
	path := filepath.Join(wrkHome, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("wrk: read config.json: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("wrk: parse config.json: %w", err)
	}
	return &cfg, nil
}

// loadCreateInterceptor returns the create.interceptor section when present.
// Missing config or missing interceptor returns (nil, nil).
func loadCreateInterceptor(wrkHome string) (*CreateInterceptor, error) {
	cfg, err := loadConfig(wrkHome)
	if err != nil {
		return nil, err
	}
	if cfg == nil || cfg.Create == nil || cfg.Create.Interceptor == nil {
		return nil, nil
	}
	return cfg.Create.Interceptor, nil
}

// loadConfigMap reads $WRK_HOME/config.json as a generic map so unknown keys
// can be preserved across management writes. Missing file returns (nil, nil).
func loadConfigMap(wrkHome string) (map[string]interface{}, error) {
	path := filepath.Join(wrkHome, "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("wrk: read config.json: %w", err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("wrk: parse config.json: %w", err)
	}
	if root == nil {
		root = map[string]interface{}{}
	}
	return root, nil
}

// saveConfigMap writes config.json with indent + trailing newline via temp+rename.
func saveConfigMap(wrkHome string, root map[string]interface{}) error {
	if root == nil {
		root = map[string]interface{}{}
	}
	path := filepath.Join(wrkHome, "config.json")
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		return fmt.Errorf("wrk: create WRK_HOME: %w", err)
	}
	data, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return fmt.Errorf("wrk: marshal config.json: %w", err)
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("wrk: write config.json: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("wrk: write config.json: %w", err)
	}
	return nil
}

// interceptorFromConfigMap extracts create.interceptor, or (nil, nil) if absent.
func interceptorFromConfigMap(root map[string]interface{}) (*CreateInterceptor, error) {
	if root == nil {
		return nil, nil
	}
	createVal, ok := root["create"]
	if !ok || createVal == nil {
		return nil, nil
	}
	createMap, ok := createVal.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("wrk: config.json create must be an object")
	}
	icVal, ok := createMap["interceptor"]
	if !ok || icVal == nil {
		return nil, nil
	}
	// Re-marshal then unmarshal into typed CreateInterceptor (handles VarValue).
	raw, err := json.Marshal(icVal)
	if err != nil {
		return nil, fmt.Errorf("wrk: marshal interceptor: %w", err)
	}
	var ic CreateInterceptor
	if err := json.Unmarshal(raw, &ic); err != nil {
		return nil, fmt.Errorf("wrk: parse create.interceptor: %w", err)
	}
	return &ic, nil
}

// setInterceptorInConfigMap sets create.interceptor, creating create if needed.
// Preserves other top-level and create.* keys.
func setInterceptorInConfigMap(root map[string]interface{}, ic *CreateInterceptor) error {
	if root == nil {
		return fmt.Errorf("wrk: nil config map")
	}
	createVal, ok := root["create"]
	var createMap map[string]interface{}
	if ok && createVal != nil {
		createMap, ok = createVal.(map[string]interface{})
		if !ok {
			return fmt.Errorf("wrk: config.json create must be an object")
		}
	} else {
		createMap = map[string]interface{}{}
		root["create"] = createMap
	}
	raw, err := json.Marshal(ic)
	if err != nil {
		return fmt.Errorf("wrk: marshal interceptor: %w", err)
	}
	var icMap map[string]interface{}
	if err := json.Unmarshal(raw, &icMap); err != nil {
		return fmt.Errorf("wrk: marshal interceptor: %w", err)
	}
	createMap["interceptor"] = icMap
	return nil
}
