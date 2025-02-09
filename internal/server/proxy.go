package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"zuyanh.go-proxy/internal/cfg"
)

type Proxy struct {
	servers []*url.URL
	index   int
	mu      sync.Mutex
}

func NewProxy(config cfg.Configuration) *Proxy {
	var servers []*url.URL
	for _, resource := range config.Resources {
		parsedURL, err := url.Parse(resource.DestinaltionURL)
		if err != nil {
			log.Fatalf("failed to parse url: %v", err)
		}

		servers = append(servers, parsedURL)
	}

	return &Proxy{servers: servers}
}

func (p *Proxy) NextServer() *url.URL {
	p.mu.Lock()
	defer p.mu.Unlock()

	server := p.servers[p.index]
	p.index = (p.index + 1) % len(p.servers) //round robin
	return server
}

func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	server := p.NextServer()
	proxy := httputil.NewSingleHostReverseProxy(server)

	r.URL.Host = server.Host
	r.URL.Scheme = server.Scheme
	r.Header.Set("X-Forwarded-Host", r.Host)

	log.Printf("proxying request to %s", server)

	proxy.ServeHTTP(w, r)
}
