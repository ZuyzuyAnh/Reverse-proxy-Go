package main

import "zuyanh.go-proxy/internal/server"

func main() {
	if err := server.Run(); err != nil {
		panic(err)
	}
}
