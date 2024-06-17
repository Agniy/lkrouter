package main

import (
	"fmt"
	"lkrouter/config"
	"lkrouter/pkg/awslogs"
	"lkrouter/pkg/transcribe"
	"lkrouter/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	cfg := config.GetConfig()

	//start transcriber workers
	transcribeWorkChan := transcribe.InitFileTranscribeWorkers()
	SetupCloseHandler(transcribeWorkChan)

	// Init AWS logs
	// ------------------------------------
	cwl, err := awslogs.GetCwl()
	if err != nil {
		fmt.Println("Can't get cwl", err)
	} else {
		go cwl.ProcessQueue()
	}

	if cwl != nil {
		cwl.Add("Lkrouter started.")
	}
	// ------------------------------------

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

func SetupCloseHandler(messageChan chan map[string]interface{}) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, os.Kill)
	go func() {
		<-c
		close(messageChan)
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}
