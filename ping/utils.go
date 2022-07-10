package ping

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// FormatIP - trim spaces and format IP.
//
// IP - the provided IP
//
// string - return "" if the input is neither valid IPv4 nor valid IPv6
//          return IPv4 in format like "192.168.9.1"
//          return IPv6 in format like "[2002:ac1f:91c5:1::bd59]"
func FormatIP(IP string) (string, error) {

	host := strings.Trim(IP, "[ ]")
	if parseIP := net.ParseIP(host); parseIP != nil {
		// valid ip
		if parseIP.To4() == nil {
			// ipv6
			host = fmt.Sprintf("[%s]", host)
		}
		return host, nil
	}
	return "", fmt.Errorf("error IP format")
}

// ParseDuration parse the t as time.Duration, it will parse t as mills when missing unit.
func ParseDuration(t string) (time.Duration, error) {
	if timeout, err := strconv.ParseInt(t, 10, 64); err == nil {
		return time.Duration(timeout) * time.Millisecond, nil
	}
	return time.ParseDuration(t)
}

// ParseAddress will try to parse addr as url.URL.
func ParseAddress(addr string) (*url.URL, error) {
	if strings.Contains(addr, "://") {
		// it maybe with scheme, try url.Parse
		return url.Parse(addr)
	}
	return url.Parse("tcp://" + addr)
}
