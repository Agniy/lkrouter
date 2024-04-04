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

	if cfg.Debug == "1" {
		fmt.Println("Server lkrouter started at none tls mode: ", httpAddr)
		errServer := srv.ListenAndServe()
		if errServer != nil {
			panic(errServer)
		}
	} else {
		//add tls certs
		//kpr, err := keyreloader.NewKeypairReloader(cfg.CertPath, cfg.KeyPath)
		//if err != nil {
		//	log.Fatal(err)
		//}
		//
		//srv.TLSConfig = &tls.Config{
		//	GetCertificate: kpr.GetCertificateFunc(),
		//}
		fmt.Println("Server lkrouter started at tls mode: ", httpAddr)
		errServer := srv.ListenAndServeTLS(cfg.CertPath, cfg.KeyPath)
		if errServer != nil {
			panic(errServer)
		}
	}

	fmt.Println("Server lkrouter started at: ", httpAddr)
}
