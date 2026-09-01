package cloudflare

import (
	"fmt"
	"io"
	"os"
)

// StartSession finds/creates a named tunnel, routes DNS, writes config, and starts cloudflared.
// It is a thin wrapper around Attach so shared tunnel names are multi-host-safe
// (managed registry + partial Stop via Detach).
func StartSession(opts SessionOptions) (*Session, error) {
	return Attach(AttachOptions{
		Domain:     opts.Domain,
		LocalURL:   opts.LocalURL,
		TunnelName: opts.TunnelName,
		ConfigDir:  opts.ConfigDir,
		Log:        opts.Log,
		Runner:     opts.Runner,
		DNSDeleter: opts.DNSDeleter,
		Teardown:   opts.Teardown,
	})
}

// PublicBaseURL returns https://<domain> (no path).
func (s *Session) PublicBaseURL() string {
	if s == nil {
		return ""
	}
	return s.publicURL
}

// Stop releases the session.
// Managed Attach sessions detach one hostname via Detach (siblings stay up;
// connector stops only when Hosts is empty). Legacy StartSession sessions still
// kill the process, remove WorkDir, and best-effort delete DNS.
func (s *Session) Stop() error {
	if s == nil {
		return nil
	}

	// Managed attach path: partial detach from shared registry.
	if s.managed {
		err := Detach(DetachOptions{
			Domain:     s.Domain,
			TunnelName: s.TunnelName,
			ConfigDir:  s.configDir,
			Log:        s.log,
			Runner:     s.runner,
			DNSDeleter: s.dnsDeleter,
			Teardown:   s.teardown,
		})
		// Session no longer owns a live process after managed detach/restart.
		s.proc = nil
		return err
	}

	var firstErr error

	if s.proc != nil {
		if err := s.proc.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.proc = nil
	}
	// runnerMode: nothing OS-level to kill

	if s.WorkDir != "" {
		if err := os.RemoveAll(s.WorkDir); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if s.dnsDeleter != nil && s.Domain != "" {
		if err := DeleteDNS(s.dnsDeleter, s.Domain); err != nil {
			fmt.Fprintf(logOrDiscard(s.log), "warning: DeleteDNS(%s): %v\n", s.Domain, err)
			// best-effort: do not fail Stop for DNS delete errors (SIGINT path stays clean)
		}
	}

	return firstErr
}

func (s *Session) cleanupPartial() error {
	if s == nil {
		return nil
	}
	if s.ownWorkDir && s.WorkDir != "" {
		return os.RemoveAll(s.WorkDir)
	}
	return nil
}

func logOrDiscard(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	return w
}
