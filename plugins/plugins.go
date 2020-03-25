package plugins

import (
	"log"
	"net"

	"github.com/spf13/viper"

	"golang.org/x/net/dns/dnsmessage"
)

//DNSHandler main func type implemented by plugins to handle DNS requests
type DNSHandler func(net.PacketConn, net.Addr, *dnsmessage.Message) error

//DNSPlugin basic plugin interface
type DNSPlugin interface {
	Name() string
	ServeDNS(DNSHandler) DNSHandler
}

var plugins []DNSPlugin

var nullHandler = func(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
	//Update response to true so handler sends a no-answer response
	req.Header.Response = true

	return nil
}

//ChainRequest chains each DNS request to each plugin as registered
func ChainRequest(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
	chain := nullHandler

	for i := len(plugins) - 1; i >= 0; i-- {
		if isPluginDisabled(plugins[i].Name()) {
			continue
		}

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

func isPluginDisabled(name string) bool {
	disabledPlugins := viper.GetStringSlice("disabled_plugins")

	for _, plugin := range disabledPlugins {
		if plugin == name {
			return true
		}
	}

	return false
}

//QuestionsToString converts the questions of a DNS request to a hashable string
func QuestionsToString(req *dnsmessage.Message) string {
	base := ""

	for _, q := range req.Questions {
		base += q.GoString()
	}

	return base
}
