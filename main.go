package main

import (
	"lkrouter/config"
	"lkrouter/router"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := config.GetConfig()
	r := router.GetRouter()

	httpPort := cfg.Port
	httpAddr := cfg.Domain
	if httpPort != "80" {
		httpAddr += ":" + httpPort
	}

	// Create server with timeout
	srv := &http.Server{
		Addr:    httpAddr,
		Handler: r,

		// set timeout due CWE-400 - Potential Slowloris Attack
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}
