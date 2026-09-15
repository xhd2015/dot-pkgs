package installplan

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Discover finds package-main candidates under moduleRoot's cmd/ and script/
// trees and merges them per the package comment preference rules.
func Discover(moduleRoot string) (*ModulePlan, error) {
	modulePath, err := readModulePath(moduleRoot)
	if err != nil {
		return nil, err
	}
	moduleName := filepath.Base(modulePath)
	if moduleName == "" || moduleName == "." || moduleName == string(filepath.Separator) {
		return nil, fmt.Errorf("invalid module path %q", modulePath)
	}

	cmdByName := make(map[string][]string)
	scriptByName := make(map[string][]string)

	if err := collectCmdMains(moduleRoot, cmdByName); err != nil {
		return nil, err
	}
	if err := collectScriptInstalls(moduleRoot, moduleName, scriptByName); err != nil {
		return nil, err
	}

	nameSet := make(map[string]struct{}, len(cmdByName)+len(scriptByName))
	for n := range cmdByName {
		nameSet[n] = struct{}{}
	}
	for n := range scriptByName {
		nameSet[n] = struct{}{}
	}
	names := make([]string, 0, len(nameSet))
	for n := range nameSet {
		names = append(names, n)
	}
	sort.Strings(names)

	items := make([]Item, 0, len(names))
	diags := make([]Diagnostic, 0)

	for _, binName := range names {
		item, more := mergeBin(binName, moduleName, cmdByName[binName], scriptByName[binName])
		diags = append(diags, more...)
		if item != nil {
			items = append(items, *item)
		}
	}

	sort.Slice(diags, func(i, j int) bool {
		if diags[i].BinName != diags[j].BinName {
			return diags[i].BinName < diags[j].BinName
		}
		return diags[i].Kind < diags[j].Kind
	})

	return &ModulePlan{
		ModuleRoot:  moduleRoot,
		ModulePath:  modulePath,
		ModuleName:  moduleName,
		Items:       items,
		Diagnostics: diags,
	}, nil
}

func mergeBin(binName, moduleName string, cmdPaths, scriptPaths []string) (*Item, []Diagnostic) {
	cmdPaths = sortedCopy(cmdPaths)
	scriptPaths = sortedCopy(scriptPaths)

	namedRel := "./script/" + binName + "/install"
	bareRel := "./script/install"

	var named string
	var bare string
	var others []string
	for _, p := range scriptPaths {
		switch p {
		case namedRel:
			named = p
		case bareRel:
			if binName == moduleName {
				bare = p
			} else {
				others = append(others, p)
			}
		default:
			others = append(others, p)
		}
	}

	var diags []Diagnostic
	if len(cmdPaths) > 1 {
		diags = append(diags, Diagnostic{
			Level:   DiagLevelWarning,
			Kind:    DiagKindAmbiguousCmd,
			BinName: binName,
			Paths:   cmdPaths,
		})
	}

	cmdSlot := ""
	if len(cmdPaths) == 1 {
		cmdSlot = cmdPaths[0]
	}

	if named != "" && len(others) > 0 {
		diags = append(diags, Diagnostic{
			Level:   DiagLevelWarning,
			Kind:    DiagKindNestedScript,
			BinName: binName,
			Paths:   append([]string{named}, others...),
		})
	}

	scriptSlot := ""
	switch {
	case named != "":
		scriptSlot = named
	case bare != "":
		scriptSlot = bare
	case len(others) == 1:
		scriptSlot = others[0]
	case len(others) > 1:
		diags = append(diags, Diagnostic{
			Level:   DiagLevelWarning,
			Kind:    DiagKindAmbiguousScript,
			BinName: binName,
			Paths:   others,
		})
	}

	if scriptSlot != "" {
		skipped := skippedLower(scriptSlot, named, bare, others, cmdSlot)
		if len(skipped) > 0 {
			diags = append(diags, Diagnostic{
				Level:   DiagLevelNotice,
				Kind:    DiagKindPreferScript,
				BinName: binName,
				Paths:   append([]string{scriptSlot}, skipped...),
			})
		}
		return &Item{
			BinName: binName,
			RelPath: scriptSlot,
			Method:  MethodGoRunInstall,
		}, diags
	}
	if cmdSlot != "" {
		return &Item{
			BinName: binName,
			RelPath: cmdSlot,
			Method:  MethodGoInstall,
		}, diags
	}
	return nil, diags
}

// skippedLower lists present lower-priority paths relative to the winner,
// excluding nested extras already reported as nested-script.
func skippedLower(winner, named, bare string, others []string, cmdSlot string) []string {
	var skipped []string
	seen := map[string]struct{}{winner: {}}
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		skipped = append(skipped, p)
	}
	// Prefer-script names the conventional lower slots, not nested extras
	// (those have their own warning).
	if winner == named {
		add(bare)
		add(cmdSlot)
		return skipped
	}
	if winner == bare {
		add(cmdSlot)
		return skipped
	}
	// Unique nested fallback: still notice over cmd.
	add(cmdSlot)
	return skipped
}

func sortedCopy(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	out := append([]string(nil), paths...)
	sort.Strings(out)
	return out
}

func collectCmdMains(moduleRoot string, byName map[string][]string) error {
	cmdRoot := filepath.Join(moduleRoot, "cmd")
	info, err := os.Stat(cmdRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	return filepath.WalkDir(cmdRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if path != cmdRoot {
			if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}
		ok, err := isPackageMainDir(path)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		rel, err := filepath.Rel(moduleRoot, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		binName := filepath.Base(path)
		byName[binName] = append(byName[binName], "./"+relSlash)
		return nil
	})
}

func collectScriptInstalls(moduleRoot, moduleName string, byName map[string][]string) error {
	scriptRoot := filepath.Join(moduleRoot, "script")
	info, err := os.Stat(scriptRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	return filepath.WalkDir(scriptRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		name := d.Name()
		if path != scriptRoot {
			if name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}
		if name != "install" {
			return nil
		}
		ok, err := isPackageMainDir(path)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		parent := filepath.Base(filepath.Dir(path))
		binName := parent
		if parent == "script" {
			if filepath.Clean(filepath.Dir(path)) == filepath.Clean(scriptRoot) {
				binName = moduleName
			}
		}
		rel, err := filepath.Rel(moduleRoot, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		byName[binName] = append(byName[binName], "./"+relSlash)
		return nil
	})
}

func isPackageMainDir(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly)
		if err != nil {
			continue
		}
		if f.Name != nil && f.Name.Name == "main" {
			return true, nil
		}
	}
	return false, nil
}

func readModulePath(moduleRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(moduleRoot, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			return fields[1], nil
		}
		if len(fields) == 1 && fields[0] == "module" {
			return "", fmt.Errorf("go.mod: module directive missing path")
		}
	}
	return "", fmt.Errorf("go.mod: no module path found")
}
