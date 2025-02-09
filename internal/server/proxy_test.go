package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func newMockServer(response string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}))
}

func boostrapServers() []*url.URL {
	backend1 := newMockServer("backend1")
	backend2 := newMockServer("backend2")
	backend3 := newMockServer("backend3")

	urls := []*url.URL{
		mustParseURL(backend1.URL),
		mustParseURL(backend2.URL),
		mustParseURL(backend3.URL),
	}

	return urls
}

func NewTestProxy() *httptest.Server {
	proxy := Proxy{
		servers: boostrapServers(),
		clients: make(map[string]*client),
	}

	proxyServer := httptest.NewServer(http.HandlerFunc(proxy.ProxyHandler))
	return proxyServer
}

func TestProxyHandler(t *testing.T) {

	proxyServer := NewTestProxy()
	defer proxyServer.Close()

	client := http.Client{}

	expectedResponse := []string{"backend1", "backend2", "backend3"}
	for i := 0; i < 3; i++ {
		response, err := client.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer response.Body.Close()

		body, _ := io.ReadAll(response.Body)
		if string(body) != expectedResponse[i] {
			t.Errorf("expected response %s, got %s", expectedResponse[i], body)
		}
	}
}

func TestRateLimit(t *testing.T) {
	proxyServer := NewTestProxy()

	client := http.Client{}

	allowed := 0
	blocked := 0

	for i := 0; i < 10; i++ {
		response, err := client.Get(proxyServer.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer response.Body.Close()

		body, _ := io.ReadAll(response.Body)

		if response.StatusCode == http.StatusTooManyRequests {
			t.Logf("Request %d blocked: %s", i+1, body)
			blocked++
		} else {
			t.Logf("Request %d allowed: %s", i+1, body)
			allowed++
		}

		time.Sleep(100 * time.Millisecond)
	}

	if blocked == 0 {
		t.Errorf("expected at least one request to be blocked")
	} else {
		t.Logf("Allowed requests: %d, Blocked requests: %d", allowed, blocked)
	}
}

func mustParseURL(plainUrl string) *url.URL {
	parsedURL, err := url.Parse(plainUrl)
	if err != nil {
		panic(err)
	}
	return parsedURL
}
