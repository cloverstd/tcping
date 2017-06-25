package ping

import (
	"fmt"
	"net"
	"time"
)

// TCPing ...
type TCPing struct {
	host string
	port int

	protocol Protocol
	counter  int
	timeout  time.Duration
	done     chan struct{}
	interval time.Duration
	result   Result
}

var _ Ping = (*TCPing)(nil)

// NewTCPing return a new TCPing
func NewTCPing(host string, port int, protocol Protocol, counter int, timeout time.Duration, interval time.Duration) *TCPing {
	tcping := TCPing{
		host:     host,
		port:     port,
		protocol: protocol,
		counter:  counter,
		timeout:  timeout,
		interval: interval,
		done:     make(chan struct{}),
	}
	tcping.result = Result{
		Pinger: &tcping,
	}
	return &tcping
}

// Host return the host of ping
func (tcping TCPing) Host() string {
	return tcping.host
}

// Port return the port of ping
func (tcping TCPing) Port() int {
	return tcping.port
}

// Protocol return the ping protocol
func (tcping TCPing) Protocol() Protocol {
	return tcping.protocol
}

// Counter return the ping counter
func (tcping TCPing) Counter() int {
	return tcping.counter
}

// Start the tcping
func (tcping *TCPing) Start() <-chan struct{} {
	go func() {
		t := time.NewTicker(tcping.interval)
		for {
			select {
			case <-t.C:
				if tcping.result.Counter >= tcping.counter && tcping.counter != 0 {
					tcping.Stop()
					return
				}
				duration, err := tcping.ping()
				if err != nil {
					fmt.Printf("Ping %s://%s:%d - failed: %s\n", tcping.protocol, tcping.host, tcping.port, err)
				} else {
					if tcping.result.MinDuration == 0 {
						tcping.result.MinDuration = duration
					}
					if tcping.result.MaxDuration == 0 {
						tcping.result.MaxDuration = duration
					}
					fmt.Printf("Ping %s://%s:%d - Connected - time=%s\n", tcping.protocol, tcping.host, tcping.port, duration)
					tcping.result.SuccessCounter++
					if duration > tcping.result.MaxDuration {
						tcping.result.MaxDuration = duration
					} else if duration < tcping.result.MinDuration {
						tcping.result.MinDuration = duration
					}
					tcping.result.TotalDuration += duration
				}
				tcping.result.Counter++
			case <-tcping.done:
				return
			}
		}
	}()
	return tcping.done
}

// Stop the tcping
func (tcping *TCPing) Stop() {
	tcping.done <- struct{}{}
}

func (tcping TCPing) ping() (time.Duration, error) {
	duration, errIfce := timeIt(func() interface{} {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", tcping.host, tcping.port), tcping.timeout)
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	})
	if errIfce != nil {
		err := errIfce.(error)
		return 0, err
	}
	return time.Duration(duration), nil
}

// Result return Result
func (tcping TCPing) Result() Result {
	return tcping.result
}
