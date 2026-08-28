package iterm2

import "github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/tabselect"

// Tab selection types/funcs live in package tabselect; aliases keep iterm2 importers stable.
type (
	TabSelKind  = tabselect.TabSelKind
	TabSelector = tabselect.TabSelector
)

const (
	TabSelAbs1  = tabselect.TabSelAbs1
	TabSelAbs0  = tabselect.TabSelAbs0
	TabSelNext  = tabselect.TabSelNext
	TabSelLeft  = tabselect.TabSelLeft
)

// ParseTabFlag parses --tab values: 1-based N, or next|right|left.
func ParseTabFlag(raw string) (TabSelector, error) {
	return tabselect.ParseTabFlag(raw)
}

// ParseTabIndexFlag parses --tab-index N (0-based).
func ParseTabIndexFlag(n int) (TabSelector, error) {
	return tabselect.ParseTabIndexFlag(n)
}

// SelectWindowTab picks a tab from st according to sel.
func SelectWindowTab(st WindowStatus, sel TabSelector) (TabStatusRow, int, error) {
	return tabselect.SelectWindowTab(st, sel)
}
