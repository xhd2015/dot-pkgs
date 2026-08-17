package client

import "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"

// SessionInfo is an alias for server session metadata.
type SessionInfo = ptywrap.SessionInfo

// ConnectOptions configures a WebSocket terminal attach.
type ConnectOptions struct {
	SessionID string
	Name      string
	Cwd       string
	// AttachMode is the ?attach_mode= query value. When set, it takes
	// precedence over AttachSnapshot. Use "attach" for an interactive
	// client that must type even if another client already holds the writer.
	AttachMode string
	// AttachSnapshot sets attach_mode=screen (live CUP snapshot; writer if
	// free, otherwise a read-only observer). Ignored when AttachMode is set.
	AttachSnapshot bool
	Wait           bool
	SkipTTYCheck   bool
	AuthToken      string
}

// AttachResult holds metadata from a completed or handshaken attach.
type AttachResult struct {
	SessionID string
	// Detached is true when the client left via Ctrl-] (detach_keep).
	// The remote PTY child is still running.
	Detached bool
}