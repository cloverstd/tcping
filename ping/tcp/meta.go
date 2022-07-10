package tcp

import (
	"fmt"
	"strings"
	"time"
)

var _ fmt.Stringer = (*Meta)(nil)

type Meta struct {
	version    int
	dnsNames   []string
	serverName string
	notBefore  time.Time
	notAfter   time.Time
}

func (m Meta) String() string {
	return fmt.Sprintf(
		"serverName=%s version=%d notBefore=%s notAfter=%s dnsNames=%s",
		m.serverName,
		m.version,
		formatTime(m.notBefore),
		formatTime(m.notAfter),
		strings.Join(m.dnsNames, ","),
	)
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
