// Package window holds iTerm2 window/tab status models and pure build/format
// helpers. Live resolve (CurrentWindowStatus) stays in package iterm2 so this
// package does not import the parent (avoids cycles with future tabselect).
package window

// CurrentLocation is the iTerm2 pane that parents the calling process.
type CurrentLocation struct {
	WindowID    string
	WindowName  string
	TabIndex    int // 1-based
	SessionID   string
	TTY         string
	SessionName string
}

// TabStatusRow is one tab line for window status output.
type TabStatusRow struct {
	Index     int
	Current   bool
	Name      string
	SessionID string
	TTY       string
}

// WindowStatus is the parent window plus all of its tabs.
type WindowStatus struct {
	WindowID        string
	WindowName      string
	CurrentTabIndex int
	Tabs            []TabStatusRow
}

// TabStatus is a summary of the parent tab only.
type TabStatus struct {
	WindowID   string
	WindowName string
	TabIndex   int
	Name       string
	SessionID  string
	TTY        string
}

// PaneRef is one session pane used when building window status.
// Field layout matches iterm2.SessionRef so callers can convert 1:1.
type PaneRef struct {
	WindowID   string
	WindowName string
	TabIndex   int // 1-based
	SessionID  string
	TTY        string
	Name       string
}
