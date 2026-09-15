// Package installplan discovers how a Go module tree builds named CLI binaries.
//
// Preference per bin name (package-main dirs only):
//
//	./script/<cmd>/install  >  ./script/install (only when cmd == module basename)
//	                        >  ./cmd/<cmd>
//
// Unique nested script/**/install paths with the same bin name are a fallback
// when the exact named script slot is missing. When the exact named slot exists,
// nested extras are warnings and are not selected.
package installplan

// Method is how the binary is built.
type Method string

const (
	MethodGoInstall    Method = "go-install"
	MethodGoRunInstall Method = "go-run-install"
)

const (
	DiagLevelNotice  = "notice"
	DiagLevelWarning = "warning"

	DiagKindPreferScript    = "prefer-script"
	DiagKindAmbiguousCmd    = "ambiguous-cmd"
	DiagKindAmbiguousScript = "ambiguous-script"
	DiagKindNestedScript    = "nested-script"
)

// Diagnostic is a non-fatal notice or warning produced while merging
// cmd/ and script/ candidates for the same BinName.
type Diagnostic struct {
	Level   string // "notice" | "warning"
	Kind    string
	BinName string
	Paths   []string // prefer-script: [winner, skipped...]; others: involved paths
}

// Item is one discovered binary build candidate.
type Item struct {
	BinName string
	RelPath string // "./cmd/foo" or "./script/foo/install"
	Method  Method // "go-install" | "go-run-install"
}

// ModulePlan is discovery for one Go module.
type ModulePlan struct {
	ModuleRoot  string
	ModulePath  string // full module path from go.mod
	ModuleName  string // basename of module path from go.mod
	RelDir      string // module root relative to scan root ("." or slash path)
	Items       []Item // sorted lexicographically by BinName
	Diagnostics []Diagnostic
}

// MultiPlan is multi-module discovery. Modules are ordered lexicographically
// by absolute ModuleRoot.
type MultiPlan struct {
	Modules []ModulePlan
}
