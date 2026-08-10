package eventbus

import "encoding/json"

// Event is the shared JSON envelope for bus notifications.
// Field names are the locked wire contract used by publishers and hubs.
type Event struct {
	ID      string          `json:"id"`
	TS      string          `json:"ts"`
	Source  string          `json:"source"`
	Type    string          `json:"type"`
	Host    string          `json:"host,omitempty"`
	Payload json.RawMessage `json:"payload"`
}
