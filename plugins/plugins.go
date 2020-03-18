package plugins

import (
	"log"
	"net"

	"golang.org/x/net/dns/dnsmessage"
)

type DNSHandler func(net.PacketConn, net.Addr, *dnsmessage.Message) error

//DNSPlugin basic plugin interface
type DNSPlugin interface {
	Name() string
	ServeDNS(DNSHandler) DNSHandler
}

var plugins []DNSPlugin

var nullHandler = func(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
	return nil
}

//ChainRequest chains each DNS request to each plugin as registered
func ChainRequest(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
	chain := nullHandler

	for i := len(plugins) - 1; i >= 0; i-- {
		chain = plugins[i].ServeDNS(chain)
	}

	return chain(conn, addr, req)
}

//Register appends a new plugin
func Register(plugin DNSPlugin) {
	plugins = append(plugins, plugin)
	log.Printf("Registered plugin: %s", plugin.Name())
}

//RegisterBefore prepends a new plugin to ensure it's run first
func RegisterBefore(plugin DNSPlugin) {
	plugins = append([]DNSPlugin{plugin}, plugins...)
	log.Printf("Registered plugin: %s", plugin.Name())
}
