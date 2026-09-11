// Package port finds a free local TCP port for short-lived servers.
package port

import (
	"fmt"
	"net"
)

// FindAvailablePort finds an available TCP port on 127.0.0.1 starting from
// startPort. It tries startPort, startPort+1, … up to maxAttempts candidates.
// Returns the port, or an error if none are free in that range.
//
// Probing 127.0.0.1 (not ":port") matches typical local UI servers and avoids
// false "free" results when only the IPv6 wildcard is free on macOS.
func FindAvailablePort(startPort int, maxAttempts int) (int, error) {
	if maxAttempts <= 0 {
		return -1, fmt.Errorf("maxAttempts must be positive")
	}
	for i := 0; i < maxAttempts; i++ {
		currentPort := startPort + i
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", currentPort))
		if err == nil {
			listener.Close()
			return currentPort, nil
		}
	}
	return -1, fmt.Errorf("could not find available port after %d attempts starting from %d", maxAttempts, startPort)
}
