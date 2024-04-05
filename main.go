package main

import (
	"fmt"
	"lkrouter/config"
	"lkrouter/router"
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

	fmt.Println("Server lkrouter started at none tls mode: ", httpAddr)
	errServer := srv.ListenAndServe()
	if errServer != nil {
		panic(errServer)
	}

	fmt.Println("Server lkrouter started at: ", httpAddr)
}
