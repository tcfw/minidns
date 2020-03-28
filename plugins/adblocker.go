package plugins

import (
	"bufio"
	"bytes"
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
	blocker := &adblocker{
		blocked:   map[string]bool{},
		whitelist: map[string]bool{},
	}

	mStore := metrics.GetMetrics()

	mStore.RegisterPluginMetric("adblocker_blacklist", promauto.NewGauge(prometheus.GaugeOpts{
		Name: "minidns_adblock_blacklist",
		Help: "Number of records in the blacklist",
	}))

	mStore.RegisterPluginMetric("adblocker_whitelist", promauto.NewGauge(prometheus.GaugeOpts{
		Name: "minidns_adblock_whitelist",
		Help: "Number of records in the whitelist",
	}))

	mStore.RegisterPluginMetric("adblocker_update_count", promauto.NewCounter(prometheus.CounterOpts{
		Name: "minidns_adblock_update_count",
		Help: "Number of times adblocker has updated black/whitelists",
	}))

	go blocker.Start()

	Register(blocker)
}

type adblocker struct {
	lock      sync.RWMutex
	blocked   map[string]bool
	whitelist map[string]bool
}

func (ab *adblocker) Name() string {
	return "ad_blocker"
}

func (ab *adblocker) ServeDNS(h DNSHandler) DNSHandler {
	return func(conn net.PacketConn, addr net.Addr, req *dnsmessage.Message) error {
		ab.lock.RLock()
		defer ab.lock.RUnlock()

		query := req.Questions[0].Name.String()

		//Remove last dot of question
		if query[len(query)-1] == '.' {
			query = query[:len(query)-1]
		}

		_, blacklisted := ab.blocked[query]
		_, whitelisted := ab.whitelist[query]

		if blacklisted && !whitelisted {
			//Simulate no answer response
			req.Header.Response = true
			return nil
		}

		return h(conn, addr, req)
	}
}

func (ab *adblocker) Start() {
	ab.update()

	timer := time.NewTicker(1 * time.Hour)
	for range timer.C {
		ab.update()
	}
}

func (ab *adblocker) update() {
	if !viper.GetBool("ready") {
		time.Sleep(50 * time.Millisecond)
		ab.update()
		return
	}

	whitelist := viper.GetStringSlice("whitelist")
	for _, domain := range whitelist {
		ab.whitelist[domain] = true
	}

	if len(ab.whitelist) > 0 {
		log.Printf("Whitelisted %d domain(s)...", len(ab.whitelist))
	}

	blacklist := viper.GetStringSlice("blacklist")
	for _, domain := range blacklist {
		ab.blocked[domain] = true
	}
	if len(ab.blocked) > 0 {
		log.Printf("Blacklisted %d domain(s)...", len(ab.blocked))
	}

	wg := sync.WaitGroup{}
	wg.Add(2)

	go ab.updateBlocklist(&wg)
	go ab.updateWhitelist(&wg)

	wg.Wait()

	metrics.GetPMetric("adblocker_update_count").(prometheus.Counter).Inc()
	metrics.GetPMetric("adblocker_blacklist").(prometheus.Gauge).Set(float64(len(ab.whitelist)))
	metrics.GetPMetric("adblocker_whitelist").(prometheus.Gauge).Set(float64(len(ab.blocked)))
}

func (ab *adblocker) updateBlocklist(wg *sync.WaitGroup) {
	blocklists := viper.GetStringSlice("blocklists")
	log.Printf("Updating block list from %d sources...\n", len(blocklists))
	ab.updateList(blocklists, ab.blocked)
	log.Printf("Updated block list: %d hosts blocked", len(ab.blocked))
	wg.Done()
}

func (ab *adblocker) updateWhitelist(wg *sync.WaitGroup) {
	whitelists := viper.GetStringSlice("whitelists")
	log.Printf("Updating whitelist from %d sources...\n", len(whitelists))
	ab.updateList(whitelists, ab.whitelist)
	log.Printf("Updated whitelist: %d whitelisted hosts", len(ab.whitelist))
	wg.Done()
}

func (ab *adblocker) updateList(list []string, hostList map[string]bool) error {
	httpClient := &http.Client{Transport: &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 10 * time.Second,
	}}

	wg := sync.WaitGroup{}
	for _, hlist := range list {
		wg.Add(1)
		go func(listEndpoint string) {
			resp, err := httpClient.Get(listEndpoint)
			if err != nil {
				log.Printf("Failed to fetch block list: %s - %s\n", listEndpoint, err)
				return
			}
			defer resp.Body.Close()

			content, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				goto err

			}

			if err := ab.processContents(content, hostList); err != nil {
				goto err
			}

		err:
			wg.Done()
			if err != nil {
				log.Printf("Failed to process block list from: %s - %s\n", listEndpoint, err)
			}
			return
		}(hlist)
	}

	wg.Wait()

	return nil
}

func (ab *adblocker) processContents(content []byte, list map[string]bool) error {
	ab.lock.Lock()
	defer ab.lock.Unlock()

	br := bytes.NewReader(content)
	s := bufio.NewScanner(br)

	for s.Scan() {
		host := s.Text()
		//Skip empty lines
		if len(host) == 0 {
			continue
		}

		//Skip comments
		if host[0] == '#' || host[0] == ';' {
			continue
		}
		list[host] = true
	}

	return nil
}
