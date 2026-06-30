package client

import "github.com/xhd2015/dot-pkgs/go-pkgs/shell/ptywrap"

// SessionInfo is an alias for server session metadata.
type SessionInfo = ptywrap.SessionInfo

// ConnectOptions configures a WebSocket terminal attach.
type ConnectOptions struct {
	SessionID      string
	Name           string
	Cwd            string
	AttachSnapshot bool
	Wait           bool
	SkipTTYCheck   bool
	AuthToken      string
}

// AttachResult holds metadata from a completed or handshaken attach.
type AttachResult struct {
	SessionID string
}