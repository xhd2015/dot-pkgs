// Package port finds a free local TCP port for short-lived servers.
package port

import (
	"fmt"
	"net"
	"time"
)

// FindAvailablePort finds an available TCP port on 127.0.0.1 starting from
// startPort. It tries startPort, startPort+1, … up to maxAttempts candidates.
// Returns the port, or an error if none are free in that range.
//
// Checks IPv4/IPv6 loopback listeners before probing a bind. A bind alone can
// succeed alongside an existing wildcard listener on macOS. This does not
// reserve the selected port; callers must still handle a later bind failure.
func FindAvailablePort(startPort int, maxAttempts int) (int, error) {
	if startPort < 1 || startPort > 65535 {
		return -1, fmt.Errorf("startPort must be between 1 and 65535")
	}
	if maxAttempts <= 0 {
		return -1, fmt.Errorf("maxAttempts must be positive")
	}
	for i := 0; i < maxAttempts; i++ {
		currentPort := startPort + i
		if currentPort > 65535 {
			break
		}
		if Listening(currentPort) {
			continue
		}
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", currentPort))
		if err == nil {
			listener.Close()
			return currentPort, nil
		}
	}
	return -1, fmt.Errorf("could not find available port after %d attempts starting from %d", maxAttempts, startPort)
}

// Listening checks both loopback families. On macOS an IPv4-specific bind can
// succeed even while a wildcard listener already accepts connections there.
func Listening(port int) bool {
	for _, host := range []string{"127.0.0.1", "::1"} {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, fmt.Sprint(port)), 150*time.Millisecond)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}
