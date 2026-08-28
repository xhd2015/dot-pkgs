package ptywrap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// RegisterAPI registers terminal HTTP and WebSocket endpoints on mux.
func RegisterAPI(mux *http.ServeMux) {
	RegisterAPIWithManager(mux, DefaultManager)
}

// RegisterAPIWithManager registers handlers using the given session manager.
func RegisterAPIWithManager(mux *http.ServeMux, mgr *Manager) {
	mux.HandleFunc("/api/terminal", func(w http.ResponseWriter, r *http.Request) {
		HandleTerminalWebSocket(w, r, mgr)
	})
	RegisterSessionAPI(mux, mgr)
}

// RegisterSessionAPI registers session-management HTTP routes.
func RegisterSessionAPI(mux *http.ServeMux, mgr *Manager) {
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		handleSessions(w, r, mgr)
	})
	mux.HandleFunc("/api/terminal/sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/terminal/sessions/")
		path = strings.TrimSuffix(path, "/")
		if strings.HasSuffix(path, "/input") {
			id := strings.TrimSuffix(path, "/input")
			if r.Method == http.MethodPost {
				handleSessionInput(w, r, mgr, id)
				return
			}
		}
		handleSessionByID(w, r, mgr)
	})
}

func handleSessionInput(w http.ResponseWriter, r *http.Request, mgr *Manager, sessionID string) {
	if sessionID == "" {
		http.Error(w, "missing session id", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body failed", http.StatusBadRequest)
		return
	}
	if err := mgr.WriteInput(sessionID, body); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleTerminalWebSocket upgrades and serves the terminal WebSocket protocol.
func HandleTerminalWebSocket(w http.ResponseWriter, r *http.Request, mgr *Manager) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	cwd := r.URL.Query().Get("cwd")
	name := r.URL.Query().Get("name")
	attachMode := r.URL.Query().Get("attach_mode")
	if name == "" {
		name = "Terminal"
	}

	var s *session
	if sessionID != "" {
		s = mgr.get(sessionID)
	}

	if s == nil {
		// Evidence for leak diagnosis: client asked to reattach but session was
		// gone (or never sent id) → createShell allocates a new PTY.
		if sessionID != "" {
			mgr.logLifecycle("reattach_miss",
				"requested_session_id", sessionID,
				"name", name,
				"cwd", cwd,
				"attach_mode", attachMode,
			)
		}
		var createErr error
		s, createErr = mgr.createShell(name, cwd)
		if createErr != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: "+createErr.Error()))
			conn.Close()
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":"session_id","session_id":"%s"}`, s.id)))
	} else {
		mgr.logLifecycle("reattach",
			"session_id", s.id,
			"requested_session_id", sessionID,
			"pid", strconv.Itoa(s.childPID()),
			"status", s.status(),
			"name", name,
			"attach_mode", attachMode,
		)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":"session_id","session_id":"%s"}`, s.id)))
	}

	ServeSessionWebSocket(conn, s.id, attachMode, mgr)
}

// ServeSessionWebSocket attaches conn to an existing session and blocks until done.
func ServeSessionWebSocket(conn *websocket.Conn, sessionID, attachMode string, mgr *Manager) {
	s := mgr.get(sessionID)
	if s == nil {
		conn.Close()
		return
	}

	role := s.claimRole(attachMode)
	s.sendRoleHandshake(conn, role)
	s.sendInitialFrame(conn, attachMode)

	if role == roleSnapshot {
		mgr.logLifecycle("attach",
			"session_id", s.id,
			"pid", strconv.Itoa(s.childPID()),
			"role", string(role),
			"attach_mode", attachMode,
			"status", s.status(),
			"snapshot_only", "true",
		)
		conn.Close()
		return
	}

	s.registerConn(conn, role)
	mgr.logLifecycle("attach",
		"session_id", s.id,
		"pid", strconv.Itoa(s.childPID()),
		"role", string(role),
		"attach_mode", attachMode,
		"status", s.status(),
		"cmd", s.commandSummary(),
	)

	type wsCloseResult struct {
		closeCode        int
		deleteOnClose    bool
		keepChildOnClose bool
	}
	wsCloseCh := make(chan wsCloseResult, 1)
	isWriter := role == roleWriter
	canWrite := isWriter || role == roleAttacher

	go func() {
		var closeCode int
		deleteOnClose := false
		// keepChildOnClose is set by {"type":"detach_keep"} so tty-watch Ctrl-]
		// (and multi-attach) can release the writer without reaping the PTY child.
		// Bare writer close still stopChild() so web create-on-connect churn does
		// not leak shells (see lifecycle leak doctests).
		keepChildOnClose := false
		defer func() {
			wsCloseCh <- wsCloseResult{
				closeCode:        closeCode,
				deleteOnClose:    deleteOnClose,
				keepChildOnClose: keepChildOnClose,
			}
		}()
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if closeErr, ok := err.(*websocket.CloseError); ok {
					closeCode = closeErr.Code
				}
				return
			}

			// detach_keep is accepted even for observers so clients can always
			// signal "leave the session running" before a normal close.
			if msgType == websocket.TextMessage {
				var msg ControlMessage
				if err := json.Unmarshal(message, &msg); err == nil && msg.Type != "" {
					switch msg.Type {
					case "resize":
						if canWrite {
							s.enqueueResize(msg.Cols, msg.Rows)
						}
						continue
					case "close_delete":
						if isWriter {
							deleteOnClose = true
						}
						continue
					case "detach_keep":
						keepChildOnClose = true
						continue
					}
				}
			}

			if !canWrite {
				continue
			}

			s.enqueueBytes(message)
		}
	}()

	// If the session already exited (e.g. re-attach after writer normal close
	// reaped the child), keep the socket open until the client disconnects so
	// scrollback already sent via sendInitialFrame can be read without a race
	// against an immediate server-side close.
	alreadyExited := false
	select {
	case <-s.done:
		alreadyExited = true
	default:
	}

	handleWriterClose := func(result wsCloseResult) {
		shouldDelete := isWriter && (result.closeCode == 4000 || result.deleteOnClose)
		action := "keep"
		reason := "not_writer"
		if shouldDelete {
			action = "remove"
			reason = "close_delete_or_4000"
		} else if isWriter && alreadyExited {
			action = "keep"
			reason = "already_exited"
		} else if isWriter && result.keepChildOnClose {
			action = "keep"
			reason = "detach_keep"
		} else if isWriter && !alreadyExited && !result.keepChildOnClose {
			action = "stop_child"
			reason = "writer_disconnect"
		}
		mgr.logLifecycle("detach",
			"session_id", s.id,
			"pid", strconv.Itoa(s.childPID()),
			"role", string(role),
			"is_writer", strconv.FormatBool(isWriter),
			"close_code", strconv.Itoa(result.closeCode),
			"delete_on_close", strconv.FormatBool(result.deleteOnClose),
			"keep_child", strconv.FormatBool(result.keepChildOnClose),
			"already_exited", strconv.FormatBool(alreadyExited),
			"action", action,
			"reason", reason,
			"cmd", s.commandSummary(),
		)
		if shouldDelete {
			mgr.remove(s.id)
			return
		}
		s.unregisterConn(conn)
		// Writer normal close (1000) without detach_keep: free the OS PTY by
		// reaping the child, but keep session metadata+scrollback listable as
		// exited so reconnect-scrollback still works (web tab close).
		// detach_keep (tty-watch Ctrl-]) only releases the writer claim.
		if isWriter && !alreadyExited && !result.keepChildOnClose {
			s.stopChild()
		}
	}

	if alreadyExited {
		result := <-wsCloseCh
		handleWriterClose(result)
		return
	}

	select {
	case <-s.done:
		// Shell exited: send a normal close frame so clients see 1000
		// rather than abnormal closure 1006 from a bare TCP tear-down.
		s.unregisterConn(conn)
		mgr.logLifecycle("detach",
			"session_id", s.id,
			"pid", strconv.Itoa(s.childPID()),
			"role", string(role),
			"is_writer", strconv.FormatBool(isWriter),
			"action", "keep",
			"reason", "child_exited_server_close",
			"cmd", s.commandSummary(),
		)
		msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		_ = conn.WriteControl(websocket.CloseMessage, msg, time.Now().Add(time.Second))
		_ = conn.Close()
	case result := <-wsCloseCh:
		handleWriterClose(result)
	}
}

func handleSessions(w http.ResponseWriter, r *http.Request, mgr *Manager) {
	switch r.Method {
	case http.MethodGet:
		page := 1
		pageSize := 20
		if p := r.URL.Query().Get("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}
		if ps := r.URL.Query().Get("page_size"); ps != "" {
			if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 1000 {
				pageSize = parsed
			}
		}
		resp := mgr.listPaginated(page, pageSize)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case http.MethodPost:
		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		command := []string{}
		if req.Command != "" {
			command = append([]string{req.Command}, req.Args...)
		}
		var s *session
		var err error
		if len(command) == 0 {
			s, err = mgr.createShell(req.Name, req.Cwd)
		} else {
			s, err = mgr.createCommand(req.Name, req.Cwd, command)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mgr.logLifecycle("rest_create",
			"session_id", s.id,
			"pid", strconv.Itoa(s.childPID()),
			"name", s.name,
			"cwd", s.cwd,
			"cmd", s.commandSummary(),
			"via", "POST /api/terminal/sessions",
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s.info(false))
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		mgr.logLifecycle("rest_delete",
			"session_id", id,
			"via", "DELETE /api/terminal/sessions",
		)
		mgr.remove(id)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleSessionByID(w http.ResponseWriter, r *http.Request, mgr *Manager) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/terminal/sessions/")
	id := strings.TrimSuffix(path, "/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	var req RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "missing name", http.StatusBadRequest)
		return
	}
	if err := mgr.rename(id, req.Name); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	s := mgr.get(id)
	if s == nil {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.info(false))
}