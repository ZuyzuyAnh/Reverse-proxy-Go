package server

import (
	"errors"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"zuyanh.go-proxy/internal/cfg"
)

var (
	errorRateLimit = errors.New("rate limit exceeded")
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Proxy struct {
	servers  []*url.URL
	index    int
	mu       sync.Mutex
	clients  map[string]*client
	duration int
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

	return &Proxy{
		servers:  servers,
		clients:  make(map[string]*client),
		duration: config.ClientDuration,
	}
}

func (p *Proxy) NextServer() *url.URL {
	p.mu.Lock()
	defer p.mu.Unlock()

	server := p.servers[p.index]
	p.index = (p.index + 1) % len(p.servers) //round robin
	return server
}

func (p *Proxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	if err := p.RateLimit(r); err != nil {
		http.Error(w, errorRateLimit.Error(), http.StatusTooManyRequests)
		return
	}

	server := p.NextServer()

	proxy := httputil.NewSingleHostReverseProxy(server)

	r.URL.Host = server.Host
	r.URL.Scheme = server.Scheme
	r.Header.Set("X-Forwarded-Host", r.Host)

	log.Printf("proxying request to %s", server)

	proxy.ServeHTTP(w, r)
}

var once sync.Once

func (p *Proxy) RateLimit(r *http.Request) error {
	once.Do(func() {
		go func() {
			for {
				time.Sleep(time.Minute)
				p.mu.Lock()
				for ip, client := range p.clients {
					if time.Since(client.lastSeen) > time.Duration(p.duration)*time.Minute {
						delete(p.clients, ip)
					}
				}
				p.mu.Unlock()
			}
		}()
	})

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return errorRateLimit
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, ok := p.clients[ip]; !ok {
		p.clients[ip] = &client{
			limiter: rate.NewLimiter(2, 4),
		}
	}

	p.clients[ip].lastSeen = time.Now()

	if !p.clients[ip].limiter.Allow() {
		return errorRateLimit
	}

	return nil
}
