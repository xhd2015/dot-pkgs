package client

import (
	"fmt"
	"strings"
)

// ResolveTarget maps an id-or-name string to exactly one session.
func ResolveTarget(c *Client, idOrName string) (*SessionInfo, error) {
	idOrName = strings.TrimSpace(idOrName)
	if idOrName == "" {
		return nil, fmt.Errorf("terminal target cannot be empty")
	}

	sessions, err := c.List()
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if session.ID == idOrName {
			s := session
			return &s, nil
		}
	}

	var matches []SessionInfo
	for _, session := range sessions {
		if session.Name == idOrName {
			matches = append(matches, session)
		}
	}

	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no terminal session found for %q", idOrName)
	case 1:
		s := matches[0]
		return &s, nil
	default:
		ids := make([]string, 0, len(matches))
		for _, match := range matches {
			ids = append(ids, match.ID)
		}
		return nil, fmt.Errorf("terminal name %q is ambiguous; matching IDs: %s", idOrName, strings.Join(ids, ", "))
	}
}