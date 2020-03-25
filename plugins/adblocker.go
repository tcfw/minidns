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

	"github.com/spf13/viper"

	"golang.org/x/net/dns/dnsmessage"
)

func init() {
	blocker := &adblocker{
		blocked: map[string]bool{},
	}
	go blocker.Start()

	Register(blocker)
}

type adblocker struct {
	lock    sync.RWMutex
	blocked map[string]bool
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

		if _, ok := ab.blocked[query]; ok {
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

	blocklists := viper.GetStringSlice("blocklists")

	log.Printf("Updating block list from %d sources...\n", len(blocklists))

	httpClient := &http.Client{Transport: &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 10 * time.Second,
	}}

	wg := sync.WaitGroup{}

	for _, list := range blocklists {
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

			if err := ab.processContents(content); err != nil {
				goto err
			}

		err:
			wg.Done()
			if err != nil {
				log.Printf("Failed to process block list from: %s - %s\n", listEndpoint, err)
			}
			return
		}(list)
	}

	wg.Wait()

	log.Printf("Updated block list: %d hosts blocked", len(ab.blocked))
}

func (ab *adblocker) processContents(content []byte) error {
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
		ab.blocked[host] = true
	}

	return nil
}
