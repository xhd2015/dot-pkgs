package ptywrap

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	mux.HandleFunc("/api/terminal/sessions", func(w http.ResponseWriter, r *http.Request) {
		handleSessions(w, r, mgr)
	})
	mux.HandleFunc("/api/terminal/sessions/", func(w http.ResponseWriter, r *http.Request) {
		handleSessionByID(w, r, mgr)
	})
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
		var createErr error
		s, createErr = mgr.createShell(name, cwd)
		if createErr != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: "+createErr.Error()))
			conn.Close()
			return
		}
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"type":"session_id","session_id":"%s"}`, s.id)))
	} else {
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
	s.attach(conn, attachMode)

	type wsCloseResult struct {
		closeCode int
	}
	wsCloseCh := make(chan wsCloseResult, 1)
	deleteOnClose := false
	go func() {
		var closeCode int
		defer func() {
			wsCloseCh <- wsCloseResult{closeCode: closeCode}
		}()
		for {
			msgType, message, err := conn.ReadMessage()
			if err != nil {
				if closeErr, ok := err.(*websocket.CloseError); ok {
					closeCode = closeErr.Code
				}
				return
			}

			if msgType == websocket.TextMessage {
				var msg ControlMessage
				if err := json.Unmarshal(message, &msg); err == nil && msg.Type != "" {
					switch msg.Type {
					case "resize":
						s.resize(msg.Cols, msg.Rows)
						continue
					case "close_delete":
						deleteOnClose = true
						continue
					}
				}
			}

			s.ptmx.Write(message)
		}
	}()

	select {
	case <-s.done:
		s.detach(conn)
		conn.Close()
	case result := <-wsCloseCh:
		shouldDelete := result.closeCode == 4000 || deleteOnClose
		if shouldDelete {
			mgr.remove(s.id)
		} else {
			s.detach(conn)
		}
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(s.info(false))
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
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

