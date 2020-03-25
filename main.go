package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/tcfw/minidns/plugins"

	"github.com/spf13/viper"
	"golang.org/x/net/dns/dnsmessage"
)

func main() {
	if viper.GetBool("use_internal_resolver") {
		setInternalResolver()
	}

	bindAddresses := viper.GetStringSlice("bind")

	for _, addr := range bindAddresses {
		conn, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", addr, viper.GetInt("port")))
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		go listenForUDPMessages(conn)
	}

	select {}
}

func listenForUDPMessages(conn net.PacketConn) error {
	log.Println("Listening for DNS requests")

	buf := make([]byte, 1024)
	for {
		n, addr, _ := conn.ReadFrom(buf)
		msg := &dnsmessage.Message{}
		if err := msg.Unpack(buf[:n]); err != nil {
			log.Printf("failed to parse DNS request: %s\n", err)
			continue
		}

		go handleUDPRequest(conn, addr, msg)
	}
}

func handleUDPRequest(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) {
	if !req.Header.Response {
		metrics.requested++
	}

	if shouldLogVerbose() {
		log.Printf("Query: %+v", req.Questions)
	}

	if err := plugins.ChainRequest(conn, addr, req); err != nil {
		metrics.failed++
		log.Printf("failed to handle DNS request: %s\n", err)
	}

	var rejected bool = false

	if !req.Header.Response || len(req.Answers) == 0 {
		metrics.rejected++
		rejected = true
	}

	bytes, _ := req.Pack()
	conn.WriteTo(bytes, addr)

	if shouldLogVeryVerbose() {
		log.Printf("Response: %+v", req)
	}

	if !rejected {
		metrics.handled++
	}
}

func setInternalResolver() {
	upstreams := viper.GetStringSlice("forwarders")

	net.DefaultResolver = &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, "udp", fmt.Sprintf("%s:53", upstreams[0]))
		},
	}
}
