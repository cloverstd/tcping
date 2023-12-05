package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/cloverstd/tcping/ping"
	"github.com/cloverstd/tcping/ping/http"
	"github.com/cloverstd/tcping/ping/tcp"
	"github.com/spf13/cobra"
)

var (
	showVersion bool
	version     string
	gitCommit   string
	counter     int
	timeout     string
	interval    string
	sigs        chan os.Signal

	httpMethod string
	httpUA     string

	dnsServer []string
)

var rootCmd = cobra.Command{
	Use:   "tcping host port",
	Short: "tcping is a tcp ping",
	Long:  "tcping is a ping over tcp connection",
	Example: `
  1. ping over tcp
	> tcping google.com
  2. ping over tcp with custom port
	> tcping google.com 443
  3. ping over http
  	> tcping http://google.com
  4. ping with URI schema
  	> tcping https://hui.lu
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("version: %s\n", version)
			fmt.Printf("git: %s\n", gitCommit)
			return
		}
		if len(args) == 0 {
			cmd.Usage()
			return
		}
		if len(args) > 2 {
			cmd.Println("invalid command arguments")
			return
		}

		url, err := ping.ParseAddress(args[0])
		if err != nil {
			fmt.Printf("%s is an invalid target.\n", args[0])
			return
		}

		defaultPort := "80"
		if port := url.Port(); port != "" {
			defaultPort = port
		} else if url.Scheme == "https" {
			defaultPort = "443"
		}
		if len(args) > 1 {
			defaultPort = args[1]
		}
		port, err := strconv.Atoi(defaultPort)
		if err != nil {
			cmd.Printf("%s is invalid port.\n", defaultPort)
			return
		}
		url.Host = ping.GetUrlHost(url.Hostname(), port)


		timeoutDuration, err := ping.ParseDuration(timeout)
		if err != nil {
			cmd.Println("parse timeout failed", err)
			cmd.Usage()
			return
		}

		intervalDuration, err := ping.ParseDuration(interval)
		if err != nil {
			cmd.Println("parse interval failed", err)
			cmd.Usage()
			return
		}

		protocol, err := ping.NewProtocol(url.Scheme)
		if err != nil {
			cmd.Println("invalid protocol", err)
			cmd.Usage()
			return
		}

		option := ping.Option{
			Timeout: timeoutDuration,
		}
		if len(dnsServer) != 0 {
			option.Resolver = &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (conn net.Conn, err error) {
					for _, addr := range dnsServer {
						ipAddr, err := ping.FormatIP(addr)
						if err != nil {
							ipAddr = addr
						}
						if conn, err = net.Dial("udp", ipAddr+":53"); err != nil {
							continue
						} else {
							return conn, nil
						}
					}
					return
				},
			}
		}
		pingFactory := ping.Load(protocol)
		p, err := pingFactory(url, &option)
		if err != nil {
			cmd.Println("load pinger failed", err)
			cmd.Usage()
			return
		}

		pinger := ping.NewPinger(os.Stdout, url, p, intervalDuration, counter)
		sigs = make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go pinger.Ping()
		select {
		case <-sigs:
		case <-pinger.Done():
		}
		pinger.Stop()
		pinger.Summarize()
	},
}

func fixProxy(proxy string, op *ping.Option) error {
	if proxy == "" {
		return nil
	}
	u, err := url.Parse(proxy)
	op.Proxy = u
	return err
}

func init() {
	rootCmd.Flags().StringVar(&httpMethod, "http-method", "GET", `Use custom HTTP method instead of GET in http mode.`)
	ua := rootCmd.Flags().String("user-agent", "tcping", `Use custom UA in http mode.`)
	meta := rootCmd.Flags().Bool("meta", false, `With meta info`)
	proxy := rootCmd.Flags().String("proxy", "", "Use HTTP proxy")

	ping.Register(ping.HTTP, func(url *url.URL, op *ping.Option) (ping.Ping, error) {
		if err := fixProxy(*proxy, op); err != nil {
			return nil, err
		}
		op.UA = *ua
		return http.New(httpMethod, url.String(), op, *meta)
	})
	ping.Register(ping.HTTPS, func(url *url.URL, op *ping.Option) (ping.Ping, error) {
		if err := fixProxy(*proxy, op); err != nil {
			return nil, err
		}
		op.UA = *ua
		return http.New(httpMethod, url.String(), op, *meta)
	})
	ping.Register(ping.TCP, func(url *url.URL, op *ping.Option) (ping.Ping, error) {
		port, err := strconv.Atoi(url.Port())
		if err != nil {
			return nil, err
		}
		return tcp.New(url.Hostname(), port, op, *meta), nil
	})
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show the version and exit.")
	rootCmd.Flags().IntVarP(&counter, "counter", "c", ping.DefaultCounter, "ping counter")
	rootCmd.Flags().StringVarP(&timeout, "timeout", "T", "1s", `connect timeout, units are "ns", "us" (or "µs"), "ms", "s", "m", "h"`)
	rootCmd.Flags().StringVarP(&interval, "interval", "I", "1s", `ping interval, units are "ns", "us" (or "µs"), "ms", "s", "m", "h"`)

	rootCmd.Flags().StringArrayVarP(&dnsServer, "dns-server", "D", nil, `Use the specified dns resolve server.`)

}

func main() {

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
