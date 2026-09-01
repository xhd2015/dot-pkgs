package cloudflare

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// OriginCertDNSDeleter deletes a hostname's DNS records using credentials from
// cloudflared's origin cert (ARGO TUNNEL TOKEN: zoneID + apiToken).
type OriginCertDNSDeleter struct {
	// CertPath empty → {DefaultConfigDir}/cert.pem
	CertPath string
	// HTTPClient nil → &http.Client{}
	HTTPClient *http.Client
	// APIBase empty → https://api.cloudflare.com/client/v4
	APIBase string
}

type argoTunnelToken struct {
	AccountID string `json:"accountID"`
	ZoneID    string `json:"zoneID"`
	APIToken  string `json:"apiToken"`
}

// NewOriginCertDNSDeleter returns a deleter that reads ~/.cloudflared/cert.pem.
func NewOriginCertDNSDeleter() *OriginCertDNSDeleter {
	return &OriginCertDNSDeleter{}
}

// DeleteHostname removes DNS records whose name matches hostname (best-effort).
func (d *OriginCertDNSDeleter) DeleteHostname(hostname string) error {
	if d == nil {
		return fmt.Errorf("DNSDeleter is nil")
	}
	host := strings.TrimSpace(strings.ToLower(hostname))
	host = strings.TrimPrefix(host, "https://")
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimSuffix(host, "/")
	if host == "" {
		return fmt.Errorf("hostname is required")
	}

	tok, err := d.loadToken()
	if err != nil {
		return err
	}
	client := d.HTTPClient
	if client == nil {
		client = &http.Client{}
	}
	base := strings.TrimSuffix(d.APIBase, "/")
	if base == "" {
		base = "https://api.cloudflare.com/client/v4"
	}

	listURL := fmt.Sprintf("%s/zones/%s/dns_records?name=%s", base, url.PathEscape(tok.ZoneID), url.QueryEscape(host))
	req, err := http.NewRequest(http.MethodGet, listURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("list DNS records for %s: %w", host, err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("list DNS records for %s: HTTP %d: %s", host, resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var list struct {
		Success bool `json:"success"`
		Result  []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"result"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(body, &list); err != nil {
		return fmt.Errorf("parse DNS list for %s: %w", host, err)
	}
	if !list.Success && len(list.Result) == 0 {
		if len(list.Errors) > 0 {
			return fmt.Errorf("list DNS records for %s: %s", host, list.Errors[0].Message)
		}
		return fmt.Errorf("list DNS records for %s: unsuccessful", host)
	}
	if len(list.Result) == 0 {
		return nil // nothing to delete
	}

	var firstErr error
	for _, rec := range list.Result {
		delURL := fmt.Sprintf("%s/zones/%s/dns_records/%s", base, url.PathEscape(tok.ZoneID), url.PathEscape(rec.ID))
		dreq, err := http.NewRequest(http.MethodDelete, delURL, nil)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		dreq.Header.Set("Authorization", "Bearer "+tok.APIToken)
		dresp, err := client.Do(dreq)
		if err != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("delete DNS %s (%s): %w", host, rec.ID, err)
			}
			continue
		}
		dbody, _ := io.ReadAll(io.LimitReader(dresp.Body, 64<<10))
		_ = dresp.Body.Close()
		if dresp.StatusCode < 200 || dresp.StatusCode >= 300 {
			if firstErr == nil {
				firstErr = fmt.Errorf("delete DNS %s (%s): HTTP %d: %s", host, rec.ID, dresp.StatusCode, strings.TrimSpace(string(dbody)))
			}
		}
	}
	return firstErr
}

func (d *OriginCertDNSDeleter) loadToken() (*argoTunnelToken, error) {
	path := strings.TrimSpace(d.CertPath)
	if path == "" {
		cfgDir, err := DefaultConfigDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(cfgDir, "cert.pem")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read origin cert %s: %w", path, err)
	}
	tok, err := parseArgoTunnelToken(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse origin cert %s: %w", path, err)
	}
	return tok, nil
}

// parseArgoTunnelToken extracts zoneID/accountID/apiToken from an ARGO TUNNEL TOKEN PEM.
func parseArgoTunnelToken(pemText string) (*argoTunnelToken, error) {
	const begin = "-----BEGIN ARGO TUNNEL TOKEN-----"
	const end = "-----END ARGO TUNNEL TOKEN-----"
	s := pemText
	i := strings.Index(s, begin)
	if i < 0 {
		return nil, fmt.Errorf("missing ARGO TUNNEL TOKEN block")
	}
	s = s[i+len(begin):]
	j := strings.Index(s, end)
	if j < 0 {
		return nil, fmt.Errorf("unterminated ARGO TUNNEL TOKEN block")
	}
	b64 := strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\n', '\r', '\t':
			return -1
		default:
			return r
		}
	}, s[:j])
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("base64: %w", err)
	}
	var tok argoTunnelToken
	if err := json.Unmarshal(decoded, &tok); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	if strings.TrimSpace(tok.ZoneID) == "" || strings.TrimSpace(tok.APIToken) == "" {
		return nil, fmt.Errorf("cert token missing zoneID or apiToken")
	}
	return &tok, nil
}
