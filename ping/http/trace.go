package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http/httptrace"
	"strings"
	"time"
)

var _ fmt.Stringer = (*Trace)(nil)

type Trace struct {
	DNSDuration time.Duration `json:"dns_duration"`

	connectStart    time.Time
	ConnectDuration time.Duration `json:"connect_duration"`

	tlsStart    time.Time
	tls         bool
	TLSDuration time.Duration `json:"tls_duration"`

	WroteRequestDuration time.Duration `json:"wrote_request_duration"`

	WaitResponseDuration time.Duration `json:"wait_response_duration"`

	BodyDuration time.Duration `json:"body_duration"`

	tlsState tls.ConnectionState

	address string
}

func (t *Trace) String() string {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("connect=%s", t.ConnectDuration))

	if t.tls {
		builder.WriteString(" ")
		builder.WriteString(fmt.Sprintf("tls=%s", t.TLSDuration))
	}

	builder.WriteString(" ")
	builder.WriteString(fmt.Sprintf("request=%s", t.WroteRequestDuration))

	builder.WriteString(" ")
	builder.WriteString(fmt.Sprintf("wait_response=%s", t.WaitResponseDuration))

	builder.WriteString(" ")
	builder.WriteString(fmt.Sprintf("response_body=%s", t.WaitResponseDuration))

	return builder.String()
}

func (t *Trace) WithTrace(ctx context.Context) context.Context {
	start := time.Now()
	return httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			start = time.Now()
		},
		DNSDone: func(info httptrace.DNSDoneInfo) {
			t.DNSDuration = time.Since(start)
		},
		ConnectStart: func(network, addr string) {
			t.connectStart = time.Now()
			t.address, _, _ = net.SplitHostPort(addr)
		},
		ConnectDone: func(network, addr string, err error) {
			t.ConnectDuration = time.Since(t.connectStart)
		},
		TLSHandshakeStart: func() {
			t.tlsStart = time.Now()
			t.tls = true
		},
		TLSHandshakeDone: func(state tls.ConnectionState, err error) {
			t.TLSDuration = time.Since(t.tlsStart)
			t.tlsState = state
		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			t.WroteRequestDuration = time.Since(start) - t.TLSDuration - t.ConnectDuration - t.DNSDuration
		},
		GotFirstResponseByte: func() {
			t.WaitResponseDuration = time.Since(start) - t.WaitResponseDuration - t.TLSDuration - t.ConnectDuration - t.DNSDuration
		},
	})
}
