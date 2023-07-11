package ping

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"math"
	"net"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"
)

var pinger = map[Protocol]Factory{}

type Factory func(url *url.URL, op *Option) (Ping, error)

func Register(protocol Protocol, factory Factory) {
	pinger[protocol] = factory
}

func Load(protocol Protocol) Factory {
	return pinger[protocol]
}

// Protocol ...
type Protocol int

func (protocol Protocol) String() string {
	switch protocol {
	case TCP:
		return "tcp"
	case HTTP:
		return "http"
	case HTTPS:
		return "https"
	}
	return "unknown"
}

const (
	// TCP is tcp protocol
	TCP Protocol = iota
	// HTTP is http protocol
	HTTP
	// HTTPS is https protocol
	HTTPS
)

// NewProtocol convert protocol string to Protocol
func NewProtocol(protocol string) (Protocol, error) {
	switch strings.ToLower(protocol) {
	case TCP.String():
		return TCP, nil
	case HTTP.String():
		return HTTP, nil
	case HTTPS.String():
		return HTTPS, nil
	}
	return 0, fmt.Errorf("protocol %s not support", protocol)
}

type Option struct {
	Timeout  time.Duration
	Resolver *net.Resolver
	Proxy    *url.URL
	UA       string
}

// Target is a ping
type Target struct {
	Protocol Protocol
	Host     string
	IP       string
	Port     int
	Proxy    string

	Counter  int
	Interval time.Duration
	Timeout  time.Duration
}

func (target Target) String() string {
	return fmt.Sprintf("%s://%s:%d", target.Protocol, target.Host, target.Port)
}

type Stats struct {
	Connected   bool                    `json:"connected"`
	Error       error                   `json:"error"`
	Duration    time.Duration           `json:"duration"`
	DNSDuration time.Duration           `json:"DNSDuration"`
	Address     string                  `json:"address"`
	Meta        map[string]fmt.Stringer `json:"meta"`
	Extra       fmt.Stringer            `json:"extra"`
}

func (s *Stats) FormatMeta() string {
	keys := make([]string, 0, len(s.Meta))
	for key := range s.Meta {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for i, key := range keys {
		builder.WriteString(key)
		builder.WriteString("=")
		builder.WriteString(s.Meta[key].String())
		if i < len(keys)-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

type Ping interface {
	Ping(ctx context.Context) *Stats
}

func NewPinger(out io.Writer, url *url.URL, ping Ping, interval time.Duration, counter int) *Pinger {
	return &Pinger{
		stopC:    make(chan struct{}),
		counter:  counter,
		interval: interval,
		out:      out,
		url:      url,
		ping:     ping,
	}
}

type Pinger struct {
	ping Ping

	stopOnce sync.Once
	stopC    chan struct{}

	out io.Writer

	url *url.URL

	interval time.Duration
	counter  int

	minDuration   time.Duration
	maxDuration   time.Duration
	totalDuration time.Duration
	total         int
	failedTotal   int
}

func (p *Pinger) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopC)
	})
}

func (p *Pinger) Done() <-chan struct{} {
	return p.stopC
}

func (p *Pinger) Ping() {
	defer p.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-p.Done()
		cancel()
	}()

	interval := DefaultInterval
	if p.interval > 0 {
		interval = p.interval
	}
	timer := time.NewTimer(1)
	defer timer.Stop()

	stop := false
	p.minDuration = time.Duration(math.MaxInt64)
	for !stop {
		select {
		case <-timer.C:
			stats := p.ping.Ping(ctx)
			p.logStats(stats)
			if p.total++; p.counter > 0 && p.total > p.counter-1 {
				stop = true
			}
			timer.Reset(interval)
		case <-p.Done():
			stop = true
		}
	}
}

func (p *Pinger) Summarize() {

	const tpl = `
Ping statistics %s
	%d probes sent.
	%d successful, %d failed.
Approximate trip times:
	Minimum = %s, Maximum = %s, Average = %s`

	_, _ = fmt.Fprintf(p.out, tpl, p.url.String(), p.total, p.total-p.failedTotal, p.failedTotal, p.minDuration, p.maxDuration, p.totalDuration/time.Duration(p.total))
}

func (p *Pinger) formatError(err error) string {
	switch err := err.(type) {
	case *url.Error:
		if err.Timeout() {
			return "timeout"
		}
		return p.formatError(err.Err)
	case net.Error:
		if err.Timeout() {
			return "timeout"
		}
		if oe, ok := err.(*net.OpError); ok {
			switch err := oe.Err.(type) {
			case *os.SyscallError:
				return err.Err.Error()
			}
		}
	default:
		if errors.Is(err, context.DeadlineExceeded) {
			return "timeout"
		}
	}
	return err.Error()
}

func (p *Pinger) logStats(stats *Stats) {
	if stats.Duration < p.minDuration {
		p.minDuration = stats.Duration
	}
	if stats.Duration > p.maxDuration {
		p.maxDuration = stats.Duration
	}
	p.totalDuration += stats.Duration
	if stats.Error != nil {
		p.failedTotal++
		if errors.Is(stats.Error, context.Canceled) {
			// ignore cancel
			return
		}
	}
	status := "Failed"
	if stats.Connected {
		status = "connected"
	}

	const colorRed = "\033[0;31m"
	const colorNone = "\033[0m"

	if stats.Error != nil {
		var colorBefore, colorAfter string
		if isTerminal(p.out) {
			colorBefore = colorRed
			colorAfter = colorNone
		} else {
			colorBefore = ""
			colorAfter = ""
		}
		_, _ = fmt.Fprintf(p.out, "%sPing %s(%s) %s(%s) - time=%s dns=%s%s", colorBefore, p.url.String(), stats.Address, status, p.formatError(stats.Error), stats.Duration, stats.DNSDuration, colorAfter)
	} else {
		_, _ = fmt.Fprintf(p.out, "Ping %s(%s) %s - time=%s dns=%s", p.url.String(), stats.Address, status, stats.Duration, stats.DNSDuration)
	}
	if len(stats.Meta) > 0 {
		_, _ = fmt.Fprintf(p.out, " %s", stats.FormatMeta())
	}
	_, _ = fmt.Fprint(p.out, "\n")
	if stats.Extra != nil {
		_, _ = fmt.Fprintf(p.out, " %s\n", strings.TrimSpace(stats.Extra.String()))
	}
}

func isTerminal(out io.Writer) bool {
	if out == nil {
		return false
	}
	if f, ok := out.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}

// Result ...
type Result struct {
	Counter        int
	SuccessCounter int
	Target         *Target

	MinDuration   time.Duration
	MaxDuration   time.Duration
	TotalDuration time.Duration
}

// Avg return the average time of ping
func (result Result) Avg() time.Duration {
	if result.SuccessCounter == 0 {
		return 0
	}
	return result.TotalDuration / time.Duration(result.SuccessCounter)
}

// Failed return failed counter
func (result Result) Failed() int {
	return result.Counter - result.SuccessCounter
}

func (result Result) String() string {
	const resultTpl = `
Ping statistics {{.Target}}
	{{.Counter}} probes sent.
	{{.SuccessCounter}} successful, {{.Failed}} failed.
Approximate trip times:
	Minimum = {{.MinDuration}}, Maximum = {{.MaxDuration}}, Average = {{.Avg}}`
	t := template.Must(template.New("result").Parse(resultTpl))
	res := bytes.NewBufferString("")
	_ = t.Execute(res, result)
	return res.String()
}
