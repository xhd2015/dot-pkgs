package port

import (
	"net"
	"testing"
)

func TestWildcardListenerIsOccupied(t *testing.T) {
	for _, network := range []string{"tcp", "tcp4", "tcp6"} {
		t.Run(network, func(t *testing.T) {
			ln, err := net.Listen(network, ":0")
			if err != nil {
				t.Skip(err)
			}
			defer ln.Close()
			p := ln.Addr().(*net.TCPAddr).Port
			if !Listening(p) {
				t.Fatal("missed wildcard listener")
			}
			if _, err := FindAvailablePort(p, 1); err == nil {
				t.Fatal("returned occupied wildcard port")
			}
		})
	}
}
