package client

import (
	"time"
)

// WaitSession blocks until the remote session exits.
func WaitSession(c *Client, sessionID string) error {
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		sessions, err := c.List()
		if err != nil {
			return err
		}
		for _, s := range sessions {
			if s.ID != sessionID {
				continue
			}
			if s.Status == "exited" {
				return nil
			}
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	return nil
}