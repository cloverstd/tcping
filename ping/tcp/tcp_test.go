package tcp_test

import (
	"context"
	tcping "github.com/cloverstd/tcping/ping"
	"github.com/cloverstd/tcping/ping/tcp"
	"net"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	ping := tcp.New("google.com", 80, &tcping.Option{}, false)
	stats := ping.Ping(context.Background())
	if !stats.Connected {
		t.Fatalf("ping failed, %s", stats.Error)
	}
}

func TestPing_Failed(t *testing.T) {

	ping := tcp.New("127.0.0.1", 1, &tcping.Option{}, false)
	stats := ping.Ping(context.Background())
	if stats.Connected {
		t.Fatalf("it should be connected refused error")
	}
}

func TestPing_IPv6(t *testing.T) {
	ln, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		t.Skipf("ipv6 is unavailable: %v", err)
	}
	defer ln.Close()

	accepted := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
		}
		close(accepted)
	}()

	addr := ln.Addr().(*net.TCPAddr)
	ping := tcp.New(addr.IP.String(), addr.Port, &tcping.Option{}, false)
	stats := ping.Ping(context.Background())
	if !stats.Connected {
		t.Fatalf("ipv6 ping failed, %s", stats.Error)
	}

	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("accept timed out")
	}
}
