# minidns

<img class="badge" tag="github.com/tcfw/minidns" src="https://goreportcard.com/badge/github.com/tcfw/minidns">

A very small caching DNS server written in Go

### Plugins
- Cache: caches known answers until TTL runs out
- Forwarder: forwards DNS questions to upstream DNS servers with a 1 second timeout per upstream endpoint
- AdBlocker: returns empty results for given host lists to essentially block ads and malicious websites