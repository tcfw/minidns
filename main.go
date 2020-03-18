package main

import (
	"log"
	"net"

	"github.com/tcfw/minidns/plugins"

	"golang.org/x/net/dns/dnsmessage"
)

func main() {
	conn, err := net.ListenPacket("udp", ":53")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	go listenForUDPMessages(conn)

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

	if err := plugins.ChainRequest(conn, addr, req); err != nil {
		metrics.failed++
		log.Printf("failed to handle DNS request: %s\n", err)
	}
	if len(req.Answers) == 0 {
		metrics.rejected++
		return
	}

	bytes, _ := req.Pack()
	conn.WriteTo(bytes, addr)

	metrics.handled++
}
