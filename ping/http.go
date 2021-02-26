package ping

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"
)

// HTTPing ...
type HTTPing struct {
	target *Target
	done   chan struct{}
	result *Result
	Method string
}

var _ Pinger = (*HTTPing)(nil)

// NewHTTPing return new HTTPing
func NewHTTPing(method string) *HTTPing {
	return &HTTPing{
		done:   make(chan struct{}),
		Method: method,
	}
}

// SetTarget ...
func (ping *HTTPing) SetTarget(target *Target) {
	ping.target = target
	if ping.result == nil {
		ping.result = &Result{Target: target}
	}
}

// Start ping
func (ping *HTTPing) Start() <-chan struct{} {
	go func() {
		t := time.NewTicker(ping.target.Interval)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				if ping.result.Counter >= ping.target.Counter && ping.target.Counter != 0 {
					ping.Stop()
					return
				}
				duration, resp, remoteAddr, err := ping.ping()
				ping.result.Counter++

				if err != nil {
					fmt.Printf("Ping %s - failed: %s\n", ping.target, err)
				} else {
					defer resp.Body.Close()
					length, _ := io.Copy(ioutil.Discard, resp.Body)
					fmt.Printf("Ping %s(%s) - %s is open - time=%s method=%s status=%d bytes=%d\n", ping.target, remoteAddr, ping.target.Protocol, duration, ping.Method, resp.StatusCode, length)
					if ping.result.MinDuration == 0 {
						ping.result.MinDuration = duration
					}
					if ping.result.MaxDuration == 0 {
						ping.result.MaxDuration = duration
					}
					ping.result.SuccessCounter++
					if duration > ping.result.MaxDuration {
						ping.result.MaxDuration = duration
					} else if duration < ping.result.MinDuration {
						ping.result.MinDuration = duration
					}
					ping.result.TotalDuration += duration
				}
			case <-ping.done:
				return
			}
		}
	}()
	return ping.done
}

// Result return ping result
func (ping *HTTPing) Result() *Result {
	return ping.result
}

// Stop the tcping
func (ping *HTTPing) Stop() {
	ping.done <- struct{}{}
}

func (ping HTTPing) ping() (time.Duration, *http.Response, string, error) {
	var resp *http.Response
	var body io.Reader
	if ping.Method == "POST" {
		body = bytes.NewBufferString("{}")
	}
	req, err := http.NewRequest(ping.Method, ping.target.String(), body)
	if err != nil {
		return 0, nil, "", err
	}

	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), "tcping")

	var remoteAddr string
	trace := &httptrace.ClientTrace{
		ConnectStart: func(network, addr string) {
			remoteAddr = addr
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	duration, errIfce := timeIt(func() interface{} {
		client := http.Client{
			Timeout: ping.target.Timeout,
		}

		proxyProco := ping.target.Proxy
		if ping.target.Proxy != "" {

			var (
				parProxy *url.URL
				dialer   proxy.Dialer
			)

			httpTransport := &http.Transport{}

			if strings.HasPrefix(proxyProco, "http") {
				parProxy, err = url.Parse(ping.target.Proxy)
				if err != nil {
					return err
				}
				httpTransport.Proxy = http.ProxyURL(parProxy)

			} else if strings.HasPrefix(proxyProco, "sock") {

				dialer, err = proxy.SOCKS5("tcp", proxyProco[strings.Index(proxyProco, "//")+2:], nil, proxy.Direct)
				if err != nil {
					return fmt.Errorf("can't connect to the proxy: %s", err)
				}

				httpTransport.DialContext = func(ctx context.Context, network string, addr string) (net.Conn, error) {
					return dialer.Dial(network, addr)
				}

			} else {
				return fmt.Errorf("%s", "please enter the correct agency procotol")
			}

			client.Transport = httpTransport
		}

		resp, err = client.Do(req)
		return err
	})
	if errIfce != nil {
		err := errIfce.(error)
		return 0, nil, "", err
	}
	return time.Duration(duration), resp, remoteAddr, nil
}
