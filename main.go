package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/cloverstd/tcping/ping"
	"github.com/spf13/cobra"
)

var (
	showVersion bool
	version     string
	gitCommit   string
	counter     int
	proxy       string
	timeout     string
	interval    string
	sigs        chan os.Signal

	httpMode bool
	httpHead bool
	httpPost bool
	httpUA   string

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
  	> tcping -H google.com
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
		host := args[0]

		url, err := ping.ParseAddress(host)
		if err != nil {
			fmt.Printf("%s is an invalid target.\n", host)
			return
		}
		defaultPort := "80"
		if len(args) > 1 {
			defaultPort = args[1]
		}
		port, err := strconv.Atoi(defaultPort)
		if err != nil {
			cmd.Printf("%s is invalid port.\n", defaultPort)
			return
		}
		url.Host = fmt.Sprintf("%s:%d", url.Hostname(), port)

		var (
			schema string
		)

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

		if httpMode {
			url.Scheme = ping.HTTP.String()
		}
		protocol, err := ping.NewProtocol(url.Scheme)
		if err != nil {
			cmd.Println("invalid protocol", err)
			cmd.Usage()
			return
		}

		if len(dnsServer) != 0 {
			ping.UseCustomDNS(dnsServer)
		}

		parseHost, _ := ping.FormatIP(host)
		if len(parseHost) <= 0 {
			parseHost = host
		}
		target := ping.Target{
			Timeout:  timeoutDuration,
			Interval: intervalDuration,
			Host:     parseHost,
			Port:     port,
			Counter:  counter,
			Proxy:    proxy,
			Protocol: protocol,
		}
		var pinger ping.Pinger
		switch protocol {
		case ping.TCP:
			pinger = ping.NewTCPing()
		case ping.HTTP, ping.HTTPS:
			var httpMethod string
			switch {
			case httpHead:
				httpMethod = "HEAD"
			case httpPost:
				httpMethod = "POST"
			default:
				httpMethod = "GET"
			}
			pinger = ping.NewHTTPing(httpMethod)
		default:
			fmt.Printf("schema: %s not support\n", schema)
			cmd.Usage()
			return
		}
		pinger.SetTarget(&target)
		pingerDone := pinger.Start()
		select {
		case <-pingerDone:
			break
		case <-sigs:
			break
		}

		fmt.Println(pinger.Result())
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show the version and exit")
	rootCmd.Flags().IntVarP(&counter, "counter", "c", 4, "ping counter")
	rootCmd.Flags().StringVar(&proxy, "proxy", "", "Use HTTP proxy")
	rootCmd.Flags().StringVarP(&timeout, "timeout", "T", "1s", `connect timeout, units are "ns", "us" (or "µs"), "ms", "s", "m", "h"`)
	rootCmd.Flags().StringVarP(&interval, "interval", "I", "1s", `ping interval, units are "ns", "us" (or "µs"), "ms", "s", "m", "h"`)

	rootCmd.Flags().BoolVarP(&httpMode, "http", "H", false, `Use "HTTP" mode. will ignore URI Schema, force to http`)
	rootCmd.Flags().BoolVar(&httpHead, "head", false, `Use HEAD instead of GET in http mode.`)
	rootCmd.Flags().BoolVar(&httpPost, "post", false, `Use POST instead of GET in http mode.`)
	rootCmd.Flags().StringVar(&httpUA, "user-agent", "tcping", `Use custom UA in http mode.`)

	rootCmd.Flags().StringArrayVarP(&dnsServer, "dns-server", "D", nil, `Use the specified dns resolve server.`)

}

func main() {
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
