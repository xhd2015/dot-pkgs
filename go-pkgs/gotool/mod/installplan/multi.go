package installplan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

// DiscoverMulti runs Discover for each module root and returns a multi-module
// plan sorted by absolute ModuleRoot. Duplicate bin names across modules are
// allowed; Lookup reports collisions for named selection.
func DiscoverMulti(moduleRoots []string) (*MultiPlan, error) {
	if len(moduleRoots) == 0 {
		return &MultiPlan{Modules: []ModulePlan{}}, nil
	}

	modules := make([]ModulePlan, 0, len(moduleRoots))
	for _, root := range moduleRoots {
		plan, err := Discover(root)
		if err != nil {
			return nil, err
		}
		modules = append(modules, *plan)
	}

	sort.Slice(modules, func(i, j int) bool {
		return absModuleRoot(modules[i].ModuleRoot) < absModuleRoot(modules[j].ModuleRoot)
	})

	return &MultiPlan{Modules: modules}, nil
}

// ResolveScanRoot returns the absolute directory from which Go module
// discovery should begin.
//
// Rules (priority):
//  1. useMain == true: workDir must be inside a git checkout; scan root is the
//     main repository path (ResolveMainRepo of ShowToplevel).
//  2. useMain == false and inside git: scan root is ShowToplevel(workDir).
//  3. Not in git: walk up from workDir looking for a go.mod; first hit is root.
//  4. No scan root: not in git and no go.mod on the walk-up → error.
func ResolveScanRoot(workDir string, useMain bool) (string, error) {
	top, err := worktree.ShowToplevel(workDir)
	if err == nil {
		if useMain {
			mainRepo, err := worktree.ResolveMainRepo(top)
			if err != nil {
				return "", fmt.Errorf("resolve main repo for installplan scan: %w", err)
			}
			return mainRepo, nil
		}
		return top, nil
	}
	if useMain {
		return "", fmt.Errorf("resolve installplan scan root with main: not inside a git work tree: %w", err)
	}
	return findModuleRootWalking(workDir)
}

// DiscoverFromWorkDir resolves the scan root from workDir, discovers every
// Go module under that root via mod/scan, and builds a multi-module plan.
//
// Zero modules under the scan root is a hard error.
func DiscoverFromWorkDir(workDir string, useMain bool) (*MultiPlan, error) {
	scanRoot, err := ResolveScanRoot(workDir, useMain)
	if err != nil {
		return nil, err
	}
	modules, err := scan.Scan(scanRoot, scan.Options{})
	if err != nil {
		return nil, fmt.Errorf("scan modules under %s: %w", scanRoot, err)
	}
	if len(modules) == 0 {
		return nil, fmt.Errorf("no go.mod modules found under %s", scanRoot)
	}
	moduleRoots := make([]string, 0, len(modules))
	relDirByAbsRoot := make(map[string]string, len(modules))
	for _, m := range modules {
		modDir := scanRoot
		if m.Dir != "." {
			modDir = filepath.Join(scanRoot, filepath.FromSlash(m.Dir))
		}
		moduleRoots = append(moduleRoots, modDir)
		relDirByAbsRoot[absModuleRoot(modDir)] = m.Dir
	}
	plan, err := DiscoverMulti(moduleRoots)
	if err != nil {
		return nil, err
	}
	for i := range plan.Modules {
		if rel, ok := relDirByAbsRoot[absModuleRoot(plan.Modules[i].ModuleRoot)]; ok {
			plan.Modules[i].RelDir = rel
		} else {
			plan.Modules[i].RelDir = relDirFromScanRoot(scanRoot, plan.Modules[i].ModuleRoot)
		}
	}
	return plan, nil
}

func relDirFromScanRoot(scanRoot, moduleRoot string) string {
	rel, err := filepath.Rel(absModuleRoot(scanRoot), absModuleRoot(moduleRoot))
	if err != nil {
		return moduleRoot
	}
	rel = filepath.ToSlash(rel)
	if rel == "" || rel == "." {
		return "."
	}
	return rel
}

func absModuleRoot(root string) string {
	abs, err := filepath.Abs(root)
	if err != nil {
		return root
	}
	return abs
}

func findModuleRootWalking(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found walking up from %s", start)
		}
		dir = parent
	}
}

// Lookup keeps only the requested bin names. Missing names, empty names, and
// the same name claimed by multiple modules are errors. Diagnostics are kept
// only for selected bins. Item order within a module follows request order.
func Lookup(plan *MultiPlan, names []string) (*MultiPlan, error) {
	if plan == nil {
		return nil, fmt.Errorf("nil install plan")
	}
	ordered, err := dedupeNames(names)
	if err != nil {
		return nil, err
	}

	type hit struct {
		modIdx  int
		itemIdx int
	}
	byName := make(map[string][]hit)
	for mi, mod := range plan.Modules {
		for ii, it := range mod.Items {
			byName[it.BinName] = append(byName[it.BinName], hit{modIdx: mi, itemIdx: ii})
		}
	}

	selected := make([][]Item, len(plan.Modules))
	diagKeep := make([]map[string]struct{}, len(plan.Modules))
	for i := range diagKeep {
		diagKeep[i] = make(map[string]struct{})
	}

	for _, name := range ordered {
		hits := byName[name]
		if len(hits) == 0 {
			if paths, ok := namedAmbiguousPaths(plan, name); ok {
				return nil, fmt.Errorf("bin %q is ambiguous (%s)", name, strings.Join(paths, ", "))
			}
			return nil, fmt.Errorf("no install candidate for %q", name)
		}
		if len(hits) > 1 {
			a := plan.Modules[hits[0].modIdx]
			b := plan.Modules[hits[1].modIdx]
			return nil, fmt.Errorf(
				"bin %q claimed by multiple modules: %s (%s) and %s (%s)",
				name, a.ModuleRoot, a.ModuleName, b.ModuleRoot, b.ModuleName,
			)
		}
		h := hits[0]
		it := plan.Modules[h.modIdx].Items[h.itemIdx]
		selected[h.modIdx] = append(selected[h.modIdx], it)
		diagKeep[h.modIdx][name] = struct{}{}
	}

	outMods := make([]ModulePlan, 0, len(plan.Modules))
	for mi, mod := range plan.Modules {
		items := selected[mi]
		if len(items) == 0 {
			continue
		}
		var diags []Diagnostic
		for _, d := range mod.Diagnostics {
			if _, ok := diagKeep[mi][d.BinName]; ok {
				diags = append(diags, d)
			}
		}
		outMods = append(outMods, ModulePlan{
			ModuleRoot:  mod.ModuleRoot,
			ModulePath:  mod.ModulePath,
			ModuleName:  mod.ModuleName,
			RelDir:      mod.RelDir,
			Items:       items,
			Diagnostics: diags,
		})
	}
	return &MultiPlan{Modules: outMods}, nil
}

func dedupeNames(names []string) ([]string, error) {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, n := range names {
		if n == "" {
			return nil, fmt.Errorf("empty name")
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("empty name list")
	}
	return out, nil
}

func namedAmbiguousPaths(plan *MultiPlan, bin string) ([]string, bool) {
	var paths []string
	seen := make(map[string]struct{})
	for _, mod := range plan.Modules {
		for _, d := range mod.Diagnostics {
			if d.BinName != bin {
				continue
			}
			if d.Kind != DiagKindAmbiguousCmd && d.Kind != DiagKindAmbiguousScript {
				continue
			}
			for _, p := range d.Paths {
				if _, ok := seen[p]; ok {
					continue
				}
				seen[p] = struct{}{}
				paths = append(paths, p)
			}
		}
	}
	if len(paths) == 0 {
		return nil, false
	}
	sort.Strings(paths)
	return paths, true
}
