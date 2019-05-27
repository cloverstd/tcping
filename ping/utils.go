package ping

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

func timeIt(f func() interface{}) (int64, interface{}) {
	startAt := time.Now()
	res := f()
	endAt := time.Now()
	return endAt.UnixNano() - startAt.UnixNano(), res
}

// UseCustomeDNS will set the dns to default DNS resolver for global
func UseCustomeDNS(dns []string) {
	resolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (conn net.Conn, err error) {
			for _, addr := range dns {
				if conn, err = net.Dial("udp", addr+":53"); err != nil {
					continue
				} else {
					return conn, nil
				}
			}
			return
		},
	}
	net.DefaultResolver = &resolver
}

// FormatIP - trim spaces and format IP
//
// IP - the provided IP
//
// string - return "" if the input is neither valid IPv4 nor valid IPv6
//          return IPv4 in format like "192.168.9.1"
//          return IPv6 in format like "[2002:ac1f:91c5:1::bd59]"
func FormatIP(IP string) string {

	host := strings.Trim(IP, "[ ]")
	if parseIP := net.ParseIP(host); parseIP != nil && parseIP.To4() == nil {
		host = fmt.Sprintf("[%s]", host)
	}

	return host
}
