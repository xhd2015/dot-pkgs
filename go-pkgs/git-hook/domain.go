package githook

import (
	"fmt"
	"net"
	"net/url"
	"os/exec"
	"strings"
)

type DomainFilter struct {
	OriginDomain        string
	ExcludeOriginDomain string
}

func ParseDomainFlag(args []string, i int, filter *DomainFilter) (bool, int, error) {
	arg := args[i]
	switch {
	case arg == "--origin-domain":
		i++
		if i >= len(args) {
			return true, i, fmt.Errorf("--origin-domain requires a value")
		}
		filter.OriginDomain = args[i]
		return true, i, nil
	case strings.HasPrefix(arg, "--origin-domain="):
		filter.OriginDomain = strings.TrimPrefix(arg, "--origin-domain=")
		return true, i, nil
	case arg == "--exclude-origin-domain":
		i++
		if i >= len(args) {
			return true, i, fmt.Errorf("--exclude-origin-domain requires a value")
		}
		filter.ExcludeOriginDomain = args[i]
		return true, i, nil
	case strings.HasPrefix(arg, "--exclude-origin-domain="):
		filter.ExcludeOriginDomain = strings.TrimPrefix(arg, "--exclude-origin-domain=")
		return true, i, nil
	default:
		return false, i, nil
	}
}

func (f *DomainFilter) Normalize() error {
	if f.OriginDomain != "" {
		domain := NormalizeDomain(f.OriginDomain)
		if domain == "" {
			return fmt.Errorf("invalid origin domain: %s", f.OriginDomain)
		}
		f.OriginDomain = domain
	}
	if f.ExcludeOriginDomain != "" {
		domain := NormalizeDomain(f.ExcludeOriginDomain)
		if domain == "" {
			return fmt.Errorf("invalid exclude origin domain: %s", f.ExcludeOriginDomain)
		}
		f.ExcludeOriginDomain = domain
	}
	return nil
}

func (f DomainFilter) ShouldRun() (bool, error) {
	if f.OriginDomain != "" {
		ok, err := OriginDomainMatches(f.OriginDomain)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	if f.ExcludeOriginDomain != "" {
		ok, err := OriginDomainMatches(f.ExcludeOriginDomain)
		if err != nil {
			return false, err
		}
		if ok {
			return false, nil
		}
	}
	return true, nil
}

func OriginDomainMatches(want string) (bool, error) {
	remote, ok, err := GitOptionalOutput("config", "--get", "remote.origin.url")
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	host := OriginHost(strings.TrimSpace(remote))
	if host == "" {
		return false, nil
	}
	return strings.EqualFold(host, want), nil
}

func OriginHost(remote string) string {
	if remote == "" {
		return ""
	}
	if u, err := url.Parse(remote); err == nil && u.Host != "" {
		return strings.ToLower(u.Hostname())
	}
	if strings.Contains(remote, "://") {
		return ""
	}

	hostPart := remote
	if at := strings.LastIndex(hostPart, "@"); at >= 0 {
		hostPart = hostPart[at+1:]
	}
	if colon := strings.Index(hostPart, ":"); colon >= 0 {
		hostPart = hostPart[:colon]
	} else if slash := strings.Index(hostPart, "/"); slash >= 0 {
		return ""
	}
	return strings.ToLower(strings.Trim(hostPart, "[]"))
}

func NormalizeDomain(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return ""
	}
	if u, err := url.Parse(domain); err == nil && u.Host != "" {
		return strings.ToLower(u.Hostname())
	}
	if host, _, err := net.SplitHostPort(domain); err == nil {
		return strings.ToLower(strings.Trim(host, "[]"))
	}
	if strings.Contains(domain, "/") {
		return ""
	}
	if colon := strings.Index(domain, ":"); colon >= 0 {
		domain = domain[:colon]
	}
	return strings.ToLower(strings.Trim(domain, "[]"))
}

func GitOutput(args ...string) (string, error) {
	output, ok, err := GitOptionalOutput(args...)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("git %s returned no output", strings.Join(args, " "))
	}
	return output, nil
}

func GitOptionalOutput(args ...string) (string, bool, error) {
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok && exitErr.ExitCode() == 1 && strings.TrimSpace(string(output)) == "" {
			return "", false, nil
		}
		return "", false, fmt.Errorf("git %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), true, nil
}
