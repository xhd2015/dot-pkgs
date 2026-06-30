package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// Client talks to a ptywrap daemon over HTTP.
type Client struct {
	BaseURL      string
	AuthToken    string
	testSessions []SessionInfo
}

// NewClient creates a client for the daemon at baseURL (e.g. http://127.0.0.1:7681).
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/")}
}

// SetTestSessions injects canned sessions for resolve-target doctests.
func (c *Client) SetTestSessions(sessions []SessionInfo) {
	c.testSessions = append([]SessionInfo(nil), sessions...)
}

// DaemonURL returns the configured daemon base URL, honoring AGENT_TERM_SERVER.
func DaemonURL() string {
	if v := strings.TrimSpace(os.Getenv("AGENT_TERM_SERVER")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return "http://127.0.0.1:7681"
}

// NewDefaultClient creates a client using AGENT_TERM_SERVER or the default listen address.
func NewDefaultClient() *Client {
	return NewClient(DaemonURL())
}

func (c *Client) List() ([]SessionInfo, error) {
	if c.testSessions != nil {
		return append([]SessionInfo(nil), c.testSessions...), nil
	}
	page := 1
	var sessions []SessionInfo
	for {
		var out struct {
			Sessions   []SessionInfo `json:"sessions"`
			TotalPages int           `json:"total_pages"`
		}
		path := fmt.Sprintf("/api/terminal/sessions?page=%d&page_size=100", page)
		if err := c.getJSON(path, &out); err != nil {
			return nil, err
		}
		sessions = append(sessions, out.Sessions...)
		if out.TotalPages <= page || len(out.Sessions) == 0 {
			break
		}
		page++
	}
	if sessions == nil {
		sessions = []SessionInfo{}
	}
	return sessions, nil
}

// Create starts a new session and returns its metadata.
func (c *Client) Create(command []string, cwd, name string) (*SessionInfo, error) {
	body := map[string]any{}
	if len(command) > 0 {
		body["command"] = command[0]
		if len(command) > 1 {
			body["args"] = command[1:]
		}
	}
	if cwd != "" {
		body["cwd"] = cwd
	}
	if name != "" {
		body["name"] = name
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/api/terminal/sessions", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, daemonConnError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, readAPIError(resp)
	}
	var info SessionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Delete removes a session by id.
func (c *Client) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, c.BaseURL+"/api/terminal/sessions?id="+url.QueryEscape(id), nil)
	if err != nil {
		return err
	}
	c.applyAuth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return daemonConnError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	return nil
}

// Rename updates a session name.
func (c *Client) Rename(id, name string) error {
	payload, _ := json.Marshal(map[string]string{"name": name})
	req, err := http.NewRequest(http.MethodPatch, c.BaseURL+"/api/terminal/sessions/"+url.PathEscape(id), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyAuth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return daemonConnError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	return nil
}

func (c *Client) applyAuth(req *http.Request) {
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}
}

func (c *Client) getJSON(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return err
	}
	c.applyAuth(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return daemonConnError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func daemonConnError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "connection refused") || strings.Contains(msg, "connect: connection refused") {
		return fmt.Errorf("agent-term serve: daemon not running (%w)", err)
	}
	return err
}

func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	snippet := strings.TrimSpace(string(body))
	if snippet == "" {
		return fmt.Errorf("request failed: %s", resp.Status)
	}
	return fmt.Errorf("request failed: %s: %s", resp.Status, snippet)
}