package tcp

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/cloverstd/tcping/ping"
	"net"
	"net/http/httptrace"
	"time"
)

var _ ping.Ping = (*Ping)(nil)

func New(host string, port int, op *ping.Option, tls bool) *Ping {
	return &Ping{
		tls:    tls,
		host:   host,
		port:   port,
		option: op,
		dialer: &net.Dialer{
			Resolver: op.Resolver,
		},
	}
}

type Ping struct {
	option *ping.Option
	host   string
	port   int
	dialer *net.Dialer
	tls    bool
}

func (p *Ping) Ping(ctx context.Context) *ping.Stats {
	timeout := ping.DefaultTimeout
	if p.option.Timeout > 0 {
		timeout = p.option.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var stats ping.Stats
	var dnsStart time.Time
	// trace dns query
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			stats.DNSDuration = time.Since(dnsStart)
		},
	})

	start := time.Now()
	var (
		conn    net.Conn
		err     error
		tlsConn *tls.Conn
		tlsErr  error
	)
	if p.tls {
		tlsConn, err = tls.DialWithDialer(p.dialer, "tcp", ping.GetUrlHost(p.host, p.port), &tls.Config{
			InsecureSkipVerify: true,
		})
		if err == nil {
			conn = tlsConn.NetConn()
		} else {
			tlsErr = err
			conn, err = p.dialer.DialContext(ctx, "tcp", ping.GetUrlHost(p.host, p.port))
		}
	} else {
		conn, err = p.dialer.DialContext(ctx, "tcp", ping.GetUrlHost(p.host, p.port))
	}
	stats.Duration = time.Since(start)
	if err != nil {
		stats.Error = err
		if oe, ok := err.(*net.OpError); ok && oe.Addr != nil {
			stats.Address = oe.Addr.String()
		}
	} else {
		stats.Connected = true
		stats.Address = conn.RemoteAddr().String()
		if tlsConn != nil && len(tlsConn.ConnectionState().PeerCertificates) > 0 {
			state := tlsConn.ConnectionState()
			stats.Extra = Meta{
				dnsNames:   state.PeerCertificates[0].DNSNames,
				serverName: state.ServerName,
				version:    int(state.Version - tls.VersionTLS10),
				notBefore:  state.PeerCertificates[0].NotBefore,
				notAfter:   state.PeerCertificates[0].NotAfter,
			}
		} else if p.tls {
			stats.Extra = bytes.NewBufferString(fmt.Sprintf("TLS handshake failed, %s", tlsErr))
		}
	}
	return &stats
}
