// Package file parses go.mod bytes and diffs two parsed files.
//
// Parse tries strict modfile.Parse, then ParseLax. Diff reports structural
// changes; callers apply policy (for example ignoring ReplaceAdded).
package file

import (
	"fmt"
	"sort"

	"golang.org/x/mod/modfile"
)

// Parse parses go.mod data. name is used in parse errors (default "go.mod").
func Parse(name string, data []byte) (*modfile.File, error) {
	if name == "" {
		name = "go.mod"
	}
	if f, err := modfile.Parse(name, data, nil); err == nil {
		return f, nil
	}
	return modfile.ParseLax(name, data, nil)
}

// Kind is one class of go.mod structural change.
type Kind int

const (
	Module Kind = iota + 1
	Go
	Toolchain
	RequireAdded
	RequireRemoved
	RequireChanged
	ExcludeAdded
	ExcludeRemoved
	ExcludeChanged
	RetractAdded
	RetractRemoved
	RetractChanged
	ReplaceAdded
	ReplaceRemoved
	ReplaceChanged
)

func (k Kind) String() string {
	switch k {
	case Module:
		return "Module"
	case Go:
		return "Go"
	case Toolchain:
		return "Toolchain"
	case RequireAdded:
		return "RequireAdded"
	case RequireRemoved:
		return "RequireRemoved"
	case RequireChanged:
		return "RequireChanged"
	case ExcludeAdded:
		return "ExcludeAdded"
	case ExcludeRemoved:
		return "ExcludeRemoved"
	case ExcludeChanged:
		return "ExcludeChanged"
	case RetractAdded:
		return "RetractAdded"
	case RetractRemoved:
		return "RetractRemoved"
	case RetractChanged:
		return "RetractChanged"
	case ReplaceAdded:
		return "ReplaceAdded"
	case ReplaceRemoved:
		return "ReplaceRemoved"
	case ReplaceChanged:
		return "ReplaceChanged"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

// Change is one structural difference between two go.mod files.
// Key is empty for Module/Go/Toolchain. Old/New are the before/after values
// (empty for added/removed as appropriate).
type Change struct {
	Kind Kind
	Key  string
	Old  string
	New  string
}

// Diff is the set of structural changes from before → after.
type Diff struct {
	Changes []Change
}

// Empty reports whether there are no structural changes.
func (d Diff) Empty() bool {
	return len(d.Changes) == 0
}

// Without returns a copy with the given kinds removed.
func (d Diff) Without(kinds ...Kind) Diff {
	if len(kinds) == 0 || len(d.Changes) == 0 {
		return Diff{Changes: append([]Change(nil), d.Changes...)}
	}
	skip := make(map[Kind]struct{}, len(kinds))
	for _, k := range kinds {
		skip[k] = struct{}{}
	}
	out := make([]Change, 0, len(d.Changes))
	for _, c := range d.Changes {
		if _, ok := skip[c.Kind]; ok {
			continue
		}
		out = append(out, c)
	}
	return Diff{Changes: out}
}

// DiffData parses both blobs (Parse) and diffs them.
func DiffData(before, after []byte) (Diff, error) {
	b, err := Parse("go.mod", before)
	if err != nil {
		return Diff{}, err
	}
	a, err := Parse("go.mod", after)
	if err != nil {
		return Diff{}, err
	}
	return DiffFiles(b, a), nil
}

// DiffFiles compares two parsed go.mod files. Nil files are treated as empty.
func DiffFiles(before, after *modfile.File) Diff {
	if before == nil {
		before = &modfile.File{}
	}
	if after == nil {
		after = &modfile.File{}
	}
	var changes []Change
	changes = append(changes, scalarChange(Module, modulePath(before), modulePath(after))...)
	changes = append(changes, scalarChange(Go, goVersion(before), goVersion(after))...)
	changes = append(changes, scalarChange(Toolchain, toolchainName(before), toolchainName(after))...)
	changes = append(changes, mapDiff(requireVersionMap(before), requireVersionMap(after), RequireAdded, RequireRemoved, RequireChanged)...)
	changes = append(changes, mapDiff(excludeMap(before), excludeMap(after), ExcludeAdded, ExcludeRemoved, ExcludeChanged)...)
	changes = append(changes, mapDiff(retractMap(before), retractMap(after), RetractAdded, RetractRemoved, RetractChanged)...)
	changes = append(changes, mapDiff(replaceMap(before), replaceMap(after), ReplaceAdded, ReplaceRemoved, ReplaceChanged)...)
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Kind != changes[j].Kind {
			return changes[i].Kind < changes[j].Kind
		}
		return changes[i].Key < changes[j].Key
	})
	return Diff{Changes: changes}
}

func scalarChange(kind Kind, old, new string) []Change {
	if old == new {
		return nil
	}
	return []Change{{Kind: kind, Old: old, New: new}}
}

func mapDiff(before, after map[string]string, added, removed, changed Kind) []Change {
	var out []Change
	for k, bv := range before {
		av, ok := after[k]
		if !ok {
			out = append(out, Change{Kind: removed, Key: k, Old: bv})
			continue
		}
		if av != bv {
			out = append(out, Change{Kind: changed, Key: k, Old: bv, New: av})
		}
	}
	for k, av := range after {
		if _, ok := before[k]; !ok {
			out = append(out, Change{Kind: added, Key: k, New: av})
		}
	}
	return out
}

func modulePath(f *modfile.File) string {
	if f.Module == nil {
		return ""
	}
	return f.Module.Mod.Path
}

func goVersion(f *modfile.File) string {
	if f.Go == nil {
		return ""
	}
	return f.Go.Version
}

func toolchainName(f *modfile.File) string {
	if f.Toolchain == nil {
		return ""
	}
	return f.Toolchain.Name
}

func requireVersionMap(f *modfile.File) map[string]string {
	out := map[string]string{}
	for _, r := range f.Require {
		if r == nil {
			continue
		}
		out[r.Mod.Path] = r.Mod.Version
	}
	return out
}

func excludeMap(f *modfile.File) map[string]string {
	out := map[string]string{}
	for _, e := range f.Exclude {
		if e == nil {
			continue
		}
		out[e.Mod.Path] = e.Mod.Version
	}
	return out
}

func retractMap(f *modfile.File) map[string]string {
	out := map[string]string{}
	for _, r := range f.Retract {
		if r == nil {
			continue
		}
		key := r.Low
		if r.High != "" && r.High != r.Low {
			key += "," + r.High
		}
		out[key] = r.Rationale
	}
	return out
}

func replaceMap(f *modfile.File) map[string]string {
	out := map[string]string{}
	for _, r := range f.Replace {
		if r == nil {
			continue
		}
		key := r.Old.Path
		if r.Old.Version != "" {
			key += "@" + r.Old.Version
		}
		val := r.New.Path
		if r.New.Version != "" {
			val += "@" + r.New.Version
		}
		out[key] = val
	}
	return out
}
