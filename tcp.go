package cachet

type TCPMonitor struct {
	// same as output from net.JoinHostPort
	// defaults to parsed config from /etc/resolv.conf when empty
	DNSServer string

	// Will be converted to FQDN
	Domain string
	Type   string
	// expected answers (regex)
	Expect []string
}
