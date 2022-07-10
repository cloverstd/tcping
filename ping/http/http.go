package http

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	pkgurl "net/url"
	"strconv"
	"time"

	"github.com/cloverstd/tcping/ping"
)

var _ ping.Ping = (*Ping)(nil)

func New(method string, url string, op *ping.Option, trace bool) (*Ping, error) {

	_, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("url or method is invalid, %w", err)
	}

	if method == "" {
		method = http.MethodGet
	}

	return &Ping{
		url:    url,
		method: method,
		trace:  trace,
		option: op,
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// disable redirect
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				Proxy: func(r *http.Request) (*pkgurl.URL, error) {
					if op.Proxy != nil {
						return op.Proxy, nil
					}
					return http.ProxyFromEnvironment(r)
				},
				DialContext: (&net.Dialer{
					Resolver: op.Resolver,
				}).DialContext,
				DisableKeepAlives: true,
				ForceAttemptHTTP2: false,
			},
		},
	}, nil
}

type Ping struct {
	client *http.Client
	trace  bool

	option *ping.Option
	method string

	url string
}

func (p *Ping) Ping(ctx context.Context) *ping.Stats {
	timeout := ping.DefaultTimeout
	if p.option.Timeout > 0 {
		timeout = p.option.Timeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	stats := ping.Stats{
		Meta: map[string]fmt.Stringer{},
	}
	trace := Trace{}
	if p.trace {
		stats.Extra = &trace
	}
	start := time.Now()
	req, err := http.NewRequestWithContext(trace.WithTrace(ctx), p.method, p.url, nil)
	if err != nil {
		stats.Error = err
		return &stats
	}
	req.Header.Set("user-agent", p.option.UA)
	resp, err := p.client.Do(req)
	stats.DNSDuration = trace.DNSDuration
	stats.Address = trace.address

	if err != nil {
		stats.Error = err
		stats.Duration = time.Since(start)
	} else {
		stats.Meta["status"] = Int(resp.StatusCode)
		stats.Connected = true
		bodyStart := time.Now()
		defer resp.Body.Close()
		n, err := io.Copy(io.Discard, resp.Body)
		trace.BodyDuration = time.Since(bodyStart)
		if n > 0 {
			stats.Meta["bytes"] = Int(n)
		}
		stats.Duration = time.Since(start)
		if err != nil {
			stats.Connected = false
			stats.Error = fmt.Errorf("read body failed, %w", err)
		}
	}
	return &stats
}

type Int int

func (i Int) String() string {
	return strconv.Itoa(int(i))
}
