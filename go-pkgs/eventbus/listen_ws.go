package eventbus

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

// ListenWS dials wsURL, reads text JSON Event frames, and invokes onEvent for each.
// It returns when ctx is cancelled/deadline exceeded or when a non-recoverable
// dial/read/decode error occurs. After a successful receive, context cancellation
// (e.g. from the callback) is a normal stop condition.
func ListenWS(ctx context.Context, wsURL string, onEvent func(Event)) error {
	if wsURL == "" {
		return context.Canceled
	}
	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(ctx, wsURL, http.Header{})
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	defer conn.Close()

	// Unblock ReadMessage when the context ends.
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-stop:
		}
	}()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		var ev Event
		if err := json.Unmarshal(data, &ev); err != nil {
			return err
		}
		if onEvent != nil {
			onEvent(ev)
		}
	}
}
