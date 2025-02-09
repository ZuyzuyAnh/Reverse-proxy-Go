package server

import (
	"fmt"
	"net/http"

	"zuyanh.go-proxy/internal/cfg"
)

func Run() error {
	config, err := cfg.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	proxy := NewProxy(*config)

	mux := http.NewServeMux()
	mux.HandleFunc("/", proxy.ProxyHandler)

	addr := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)

	err = http.ListenAndServe(addr, mux)
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}

	return nil
}
