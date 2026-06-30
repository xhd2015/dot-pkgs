package client

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// WaitSession blocks until the remote session exits or the WebSocket closes.
func WaitSession(c *Client, sessionID string) error {
	wsURL, err := terminalWebSocketURL(c.BaseURL, ConnectOptions{SessionID: sessionID, AuthToken: c.AuthToken})
	if err != nil {
		return err
	}
	header := http.Header{}
	if c.AuthToken != "" {
		header.Set("Authorization", "Bearer "+c.AuthToken)
	}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return err
	}
	defer conn.Close()

	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				continue
			}
			return nil
		}
		if strings.Contains(string(msg), "[Terminal exited]") {
			return nil
		}
	}
	return nil
}