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

	httpAddr := ":" + cfg.Port

	// Create server with timeout
	srv := &http.Server{
		Addr:    httpAddr,
		Handler: r,

		// set timeout due CWE-400 - Potential Slowloris Attack
		ReadHeaderTimeout: 5 * time.Second,
	}

	fmt.Println("Server lkrouter started: ", httpAddr)
	errServer := srv.ListenAndServe()
	if errServer != nil {
		panic(errServer)
	}
}
