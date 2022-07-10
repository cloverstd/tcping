package ping_test

import (
	"bytes"
	"context"
	"fmt"
	tcping "github.com/cloverstd/tcping/ping"
	"net/url"
	"testing"
	"time"
)

type PingHandler func(ctx context.Context) *tcping.Stats

func (ph PingHandler) Ping(ctx context.Context) *tcping.Stats {
	return ph(ctx)
}

type String string

func (s String) String() string {
	return string(s)
}

func TestPinger(t *testing.T) {
	u, _ := url.Parse("https://hui.lu")
	var buf bytes.Buffer
	pinger := tcping.NewPinger(&buf, u, PingHandler(func(ctx context.Context) *tcping.Stats {
		return &tcping.Stats{
			Address:     "127.0.0.1:443",
			Connected:   true,
			Duration:    time.Second,
			DNSDuration: time.Millisecond * 8,
			Meta: map[string]fmt.Stringer{
				"status": String("200"),
				"byte":   String("64974"),
			},
			Extra: String("tls: 1.3"),
		}
	}), time.Second, 2)
	pinger.Ping()
	pinger.Summarize()
	fmt.Println(buf.String())
}
