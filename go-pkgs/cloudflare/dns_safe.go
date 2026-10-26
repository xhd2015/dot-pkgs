package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OriginCertDNSClient uses the existing cloudflared account credentials, without
// ever returning those credentials to an application client.
type OriginCertDNSClient struct {
	CertPath   string
	HTTPClient *http.Client
	APIBase    string
}

type DNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
}

func (c *OriginCertDNSClient) call(ctx context.Context, method, path string, body any, out any) error {
	tok, err := (&OriginCertDNSDeleter{CertPath: c.CertPath}).loadToken()
	if err != nil {
		return err
	}
	base := strings.TrimRight(c.APIBase, "/")
	if base == "" {
		base = "https://api.cloudflare.com/client/v4"
	}
	var input io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		input = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, base+strings.ReplaceAll(path, "{zone}", url.PathEscape(tok.ZoneID)), input)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok.APIToken)
	req.Header.Set("Content-Type", "application/json")
	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Cloudflare API: %w", err)
	}
	defer resp.Body.Close()
	var envelope struct {
		Success bool            `json:"success"`
		Result  json.RawMessage `json:"result"`
		Errors  []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&envelope); err != nil {
		return fmt.Errorf("Cloudflare API returned an invalid response (HTTP %d)", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || !envelope.Success {
		detail := "request failed"
		if len(envelope.Errors) > 0 {
			detail = envelope.Errors[0].Message
		}
		return fmt.Errorf("Cloudflare API HTTP %d: %s", resp.StatusCode, detail)
	}
	if out != nil {
		return json.Unmarshal(envelope.Result, out)
	}
	return nil
}

// VerifyDomain checks the login's actual zone access (a certificate existing on
// disk alone is not proof of a usable account). Subdomains of that zone work.
func (c *OriginCertDNSClient) VerifyDomain(ctx context.Context, domain string) error {
	var zone struct {
		Name string `json:"name"`
	}
	if err := c.call(ctx, http.MethodGet, "/zones/{zone}", nil, &zone); err != nil {
		return err
	}
	domain = strings.ToLower(strings.TrimSuffix(strings.TrimSpace(domain), "."))
	zone.Name = strings.ToLower(zone.Name)
	if domain != zone.Name && !strings.HasSuffix(domain, "."+zone.Name) {
		return fmt.Errorf("domain %q is not in the authenticated Cloudflare zone %q", domain, zone.Name)
	}
	return nil
}

func (c *OriginCertDNSClient) Records(ctx context.Context, hostname string) ([]DNSRecord, error) {
	var records []DNSRecord
	err := c.call(ctx, http.MethodGet, "/zones/{zone}/dns_records?name="+url.QueryEscape(hostname), nil, &records)
	return records, err
}

// EnsureTunnelCNAME creates, but NEVER overwrites, DNS. Existing records are
// reusable only when they already point to this dedicated tunnel. A concurrent
// creator is checked after a failed create instead of being overwritten.
func (c *OriginCertDNSClient) EnsureTunnelCNAME(ctx context.Context, hostname, tunnelID string) error {
	target := tunnelID + ".cfargotunnel.com"
	check := func(records []DNSRecord) error {
		for _, r := range records {
			if !strings.EqualFold(r.Name, hostname) || r.Type != "CNAME" || !strings.EqualFold(strings.TrimSuffix(r.Content, "."), target) || !r.Proxied {
				return fmt.Errorf("hostname %s already belongs to another service; no DNS record was overwritten", hostname)
			}
		}
		return nil
	}
	records, err := c.Records(ctx, hostname)
	if err != nil {
		return err
	}
	if len(records) > 0 {
		return check(records)
	}
	record := DNSRecord{Type: "CNAME", Name: hostname, Content: target, Proxied: true}
	if err := c.call(ctx, http.MethodPost, "/zones/{zone}/dns_records", record, nil); err != nil {
		records, retryErr := c.Records(ctx, hostname)
		if retryErr == nil && len(records) > 0 && check(records) == nil {
			return nil
		}
		return err
	}
	return nil
}
