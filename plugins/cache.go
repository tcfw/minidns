package plugins

import (
	"net"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tcfw/minidns/metrics"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	gcDuration = 1 * time.Minute
)

func init() {
	cr := &cacheResolver{
		cache: map[string]cacheResources{},
	}

	metrics.GetMetrics().RegisterPluginMetric("cache_records", promauto.NewGauge(prometheus.GaugeOpts{
		Name: "minidns_cache_count",
		Help: "Number of records in cache",
	}))

	go cr.StartGC()

	RegisterBefore(cr)
}

type cacheResources struct {
	expires time.Time
	created time.Time
	answers []dnsmessage.Resource
}

type cacheResolver struct {
	lock  sync.RWMutex
	cache map[string]cacheResources
}

func (cr *cacheResolver) Name() string {
	return "cache_resolver"
}

func (cr *cacheResolver) ServeDNS(h DNSHandler) DNSHandler {
	return func(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {

		cacheKey := QuestionsToString(req)

		if !req.Header.Response {
			cr.lock.RLock()
			cached, ok := cr.cache[cacheKey]
			cr.lock.RUnlock()
			if ok {
				if time.Now().After(cached.expires) {
					cr.lock.Lock()
					delete(cr.cache, cacheKey)
					cr.lock.Unlock()
				} else {
					req.Response = true
					for _, ans := range cached.answers {
						ans.Header.TTL = uint32(time.Until(cached.expires).Seconds())
						req.Answers = append(req.Answers, ans)
					}
					return nil
				}
			}
		}

		err := h(conn, addr, req)

		if req.Header.Response && len(req.Answers) > 0 {
			cr.lock.Lock()
			cr.cache[cacheKey] = cacheResources{
				created: time.Now(),
				expires: time.Now().Add(time.Duration(req.Answers[0].Header.TTL) * time.Second),
				answers: req.Answers,
			}
			cr.lock.Unlock()
		}

		return err
	}
}

func (cr *cacheResolver) StartGC() {
	timer := time.NewTicker(gcDuration)

	for range timer.C {
		cr.lock.Lock()

		for k, v := range cr.cache {
			if time.Now().After(v.expires) {
				delete(cr.cache, k)
			}
		}
		metrics.GetPMetric("cache_records").(prometheus.Gauge).Set(float64(len(cr.cache)))

		cr.lock.Unlock()
	}

}
