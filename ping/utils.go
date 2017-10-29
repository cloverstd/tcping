package ping

import (
	"context"
	"net"
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
