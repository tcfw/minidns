package plugins

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/net/dns/dnsmessage"
)

func init() {
	Register(&forwardResolver{
		wsm: sync.Map{},
	})
}

const (
	upstreamTimeout = 1 * time.Second
)

type waitResponse struct {
	msg *dnsmessage.Message
}

type forwardResolver struct {
	mu  sync.RWMutex
	wsm sync.Map
}

func (forwarder *forwardResolver) Name() string {
	return "Forward Resolver"
}

func (forwarder *forwardResolver) ServeDNS(h DNSHandler) DNSHandler {
	return func(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
		if req.Header.Response {
			forwarder.handleResponse(req)
		} else {
			forwarder.forwardAndWait(conn, addr, req)
		}

		return h(conn, addr, req)
	}
}

func (forwarder *forwardResolver) handleResponse(msg *dnsmessage.Message) {
	if !msg.Header.Response {
		return
	}

	//fanout to each waiter
	forwarder.wsm.Range(func(k interface{}, waiter interface{}) bool {
		go func() { waiter.(chan waitResponse) <- waitResponse{msg: msg} }()
		return true
	})
}

func (forwarder *forwardResolver) forwardAndWait(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) {
	upstreams := viper.GetStringSlice("forwarders")
	var upstreamAttempt int

	var answers []dnsmessage.Resource

upstreamL:
	for upstreamAttempt < len(upstreams) {
		done := make(chan []dnsmessage.Resource)
		go func() {
			timeout := time.Now().Add(5 * time.Second)
			waitKey := fmt.Sprintf("%d", req.ID)
			defer func() {
				forwarder.wsm.Delete(waitKey)
			}()

			waiter := make(chan waitResponse)
			forwarder.wsm.Store(waitKey, waiter)

			for {
				if time.Now().After(timeout) {
					return
				}

				response := <-waiter
				if !response.msg.Header.Response {
					continue
				}

				if response.msg.Questions[0] == req.Questions[0] {
					done <- response.msg.Answers
					return
				}
			}
		}()

		addr := &net.UDPAddr{
			IP:   net.ParseIP(upstreams[upstreamAttempt]),
			Port: 53,
		}

		forwarder.forwardOne(conn, addr, req)

		select {
		case upstreamAnswer := <-done:
			answers = upstreamAnswer
			break upstreamL
		case <-time.After(upstreamTimeout):
			close(done)
			upstreamAttempt++
			log.Println("upstream timed out")
		}
	}

	req.Header.Response = true
	if len(answers) > 0 {
		req.Answers = append(req.Answers, answers...)
	}
}

func (forwarder *forwardResolver) forwardOne(conn net.PacketConn, addr *net.UDPAddr, req *dnsmessage.Message) {
	bytes, _ := req.Pack()
	n, err := conn.WriteTo(bytes, addr)
	if err != nil || n == 0 {
		log.Printf("failed to forward request: %s\n", err)
	}
}
