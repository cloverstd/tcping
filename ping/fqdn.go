package ping

import "net"

// GetIP ...
func GetIP(hostname string) string {
	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			return ipv4.String()
		}
	}
	return ""
}
