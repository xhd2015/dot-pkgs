package ptywrap

import (
	"io"
	"log"
	"strings"
)

// Lifecycle logging is opt-in: set Manager.LifecycleLog to an io.Writer.
// Nil (the default) discards all events — safe for TUI hosts that share a
// terminal with the process. Callers that want diagnostics (e.g. ai-critic
// appending to ai-critic-server.log) must wire the writer explicitly.
//
// Prefix: [ptywrap]
// Format: key=value pairs on one line (grep-friendly).
//
// Events:
//   create          - new PTY session registered (shell or command)
//   reattach_miss   - client sent session_id but manager had no session → createShell
//   reattach        - client reattached to existing session
//   attach          - WS claimed a role on a session
//   detach          - WS closed; includes action taken (remove|stop_child|keep)
//   stop_child      - OS PTY child reaped (master closed + kill)
//   remove          - session deleted from manager
//   shell_exit      - child process exited on its own (readLoop EOF)
//   rest_create     - POST /api/terminal/sessions
//   rest_delete     - DELETE /api/terminal/sessions

func writeLifecycle(w io.Writer, event string, fields ...string) {
	if w == nil || event == "" {
		return
	}
	var b strings.Builder
	b.WriteString("event=")
	b.WriteString(event)
	for i := 0; i+1 < len(fields); i += 2 {
		b.WriteByte(' ')
		b.WriteString(fields[i])
		b.WriteByte('=')
		b.WriteString(quoteLifecycleField(fields[i+1]))
	}
	log.New(w, "[ptywrap] ", log.LstdFlags|log.Lmsgprefix).Print(b.String())
}

func (m *Manager) logLifecycle(event string, fields ...string) {
	if m == nil {
		return
	}
	writeLifecycle(m.LifecycleLog, event, fields...)
}

func (s *session) logLifecycle(event string, fields ...string) {
	if s == nil {
		return
	}
	writeLifecycle(s.lifecycleLog, event, fields...)
}

func quoteLifecycleField(v string) string {
	if v == "" {
		return `""`
	}
	if strings.ContainsAny(v, " \t\"=") {
		return `"` + strings.ReplaceAll(v, `"`, `'`) + `"`
	}
	return v
}

func (s *session) childPID() int {
	if s == nil || s.cmd == nil || s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}

func (s *session) commandSummary() string {
	if s == nil || len(s.command) == 0 {
		return ""
	}
	// Cap length so PATH-heavy bash -lc lines stay readable.
	joined := strings.Join(s.command, " ")
	const max = 160
	if len(joined) > max {
		return joined[:max] + "…"
	}
	return joined
}

func (m *Manager) sessionCounts() (total, running int) {
	if m == nil {
		return 0, 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		total++
		if !s.exited {
			running++
		}
	}
	return total, running
}
