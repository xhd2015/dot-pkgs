package reclaim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// iosRuntime is one Simulator runtime image from simctl list --json.
type iosRuntime struct {
	Name      string
	Version   string
	Build     string
	SizeBytes int64
}

// iosInventory is the measured iOS Simulator runtime set.
type iosInventory struct {
	Runtimes  []iosRuntime
	TotalSize int64
	// ToolMissing is true when xcrun could not be resolved.
	ToolMissing bool
	// Err is a non-missing tool failure (parse/run), if any.
	Err error
}

// measureIOSSim inventories runtimes via `xcrun simctl runtime list --json`.
// Sums each top-level map value's sizeBytes. Empty {} → 0.
func measureIOSSim(opts Options) iosInventory {
	out, err := runExternalToolOutput(opts, "xcrun", "simctl", "runtime", "list", "--json")
	if err != nil {
		inv := iosInventory{Err: err}
		if isToolMissing(err) {
			inv.ToolMissing = true
		}
		return inv
	}
	total, runtimes, parseErr := parseIOSRuntimeListJSON(out)
	if parseErr != nil {
		return iosInventory{Err: parseErr}
	}
	return iosInventory{Runtimes: runtimes, TotalSize: total}
}

// parseIOSRuntimeListJSON parses simctl runtime list --json object map.
// Shape: UUID/id → { "sizeBytes": N, "name": "...", "version": "...", "build": "..." }.
func parseIOSRuntimeListJSON(data []byte) (int64, []iosRuntime, error) {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("{}")) {
		return 0, nil, nil
	}
	var m map[string]struct {
		Name      string `json:"name"`
		Version   string `json:"version"`
		Build     string `json:"build"`
		SizeBytes int64  `json:"sizeBytes"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return 0, nil, fmt.Errorf("parse simctl runtime list: %w", err)
	}
	var total int64
	list := make([]iosRuntime, 0, len(m))
	for id, v := range m {
		total += v.SizeBytes
		name := v.Name
		if name == "" {
			name = id
		}
		list = append(list, iosRuntime{
			Name:      name,
			Version:   v.Version,
			Build:     v.Build,
			SizeBytes: v.SizeBytes,
		})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Name != list[j].Name {
			return list[i].Name < list[j].Name
		}
		return list[i].Build < list[j].Build
	})
	return total, list, nil
}

// applyIOSSimDelete runs `xcrun simctl runtime delete all`.
func applyIOSSimDelete(opts Options) error {
	return runExternalTool(opts, "xcrun", "simctl", "runtime", "delete", "all")
}

// remeasureIOSAfterDelete remeasures list size, polling briefly until empty or timeout.
// Stubs return empty immediately; real simctl may lag.
func remeasureIOSAfterDelete(opts Options) int64 {
	const (
		maxWait = 2 * time.Second
		step    = 50 * time.Millisecond
	)
	deadline := time.Now().Add(maxWait)
	var last int64
	for {
		inv := measureIOSSim(opts)
		last = inv.TotalSize
		if inv.ToolMissing || inv.Err != nil || last == 0 {
			return last
		}
		if time.Now().After(deadline) {
			return last
		}
		time.Sleep(step)
	}
}
