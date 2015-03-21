package system

import (
	"net"
	"os"
)

// GetHostname returns id of the current system
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil || len(hostname) == 0 {
		addrs, err := net.InterfaceAddrs()

		if err != nil {
			return "unknown"
		}

		for _, addr := range addrs {
			return addr.String()
		}
	}

	return hostname
}
