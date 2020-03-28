package plugins

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
	"github.com/tcfw/minidns/metrics"
	"golang.org/x/net/dns/dnsmessage"
)

func init() {
	metrics.GetMetrics().RegisterPluginMetric("doh_forwader_latency", promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "minidns_doh_forwader_query",
		Help:    "Duration of time to get response from forwarder",
		Buckets: prometheus.LinearBuckets(1, 2, 15),
	}))

	Register(newDOHForwardResolver())
}

func newDOHForwardResolver() *dohForwardResolver {
	return &dohForwardResolver{
		dohClient: http.Client{
			Transport: &http.Transport{
				MaxIdleConns:    20,
				IdleConnTimeout: 5 * time.Minute,
			},
		},
	}
}

type dohForwardResolver struct {
	mu        sync.RWMutex
	dohClient http.Client
}

func (forwarder *dohForwardResolver) Name() string {
	return "doh_forward_resolver"
}

func (forwarder *dohForwardResolver) ServeDNS(h DNSHandler) DNSHandler {
	return func(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
		if err := forwarder.forwardAndWait(conn, addr, req); err != nil {
			return err
		}

		return h(conn, addr, req)
	}
}

func (forwarder *dohForwardResolver) forwardAndWait(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
	var upstreamAttempt int
	var answers []dnsmessage.Resource

	errCh := make(chan error)
	upstreams := viper.GetStringSlice("doh_forwarders")

	sTime := time.Now()

upstreamL:
	for upstreamAttempt < len(upstreams) {
		done := make(chan []dnsmessage.Resource)
		go func() {
			reqBytes, _ := req.Pack()
			dohURL := fmt.Sprintf("https://%s/dns-query", upstreams[upstreamAttempt])
			req, _ := http.NewRequest("POST", dohURL, bytes.NewBuffer(reqBytes))
			req.Header.Add("accept", "application/dns-message")
			req.Header.Add("content-type", "application/dns-message")
			resp, err := forwarder.dohClient.Do(req)
			if err != nil {
				errCh <- err
				return
			}
			if resp.StatusCode != 200 {
				errCh <- fmt.Errorf("failed to fetch dns response - status code: %d", resp.StatusCode)
				return
			}
			if resp.Header.Get("content-type") != "application/dns-message" {
				errCh <- fmt.Errorf("unknown responses type: %s", resp.Header.Get("content-type"))
				return
			}

			defer resp.Body.Close()
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			respReq := &dnsmessage.Message{}
			respReq.Unpack(bodyBytes)
			if !respReq.Header.Response {
				errCh <- fmt.Errorf("response from DoHs not a response")
				return
			}
			if len(respReq.Answers) > 0 {
				done <- respReq.Answers
			}
		}()

		select {
		case upstreamAnswer := <-done:
			answers = upstreamAnswer
			break upstreamL
		case err := <-errCh:
			return err
		case <-time.After(upstreamTimeout):
			upstreamAttempt++
			log.Println("upstream timed out")
		}
	}

	metrics.GetPMetric("doh_forwader_latency").(prometheus.Histogram).Observe(float64(time.Since(sTime).Milliseconds()))

	req.Header.Response = true
	if len(answers) > 0 {
		req.Answers = append(req.Answers, answers...)
	}

	return nil
}
