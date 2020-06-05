package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var debugMode = false

var db = newDatabase()

// Start function sets up the url mapping and launches the HTTP
// server to listen on port 8080
func Start() {
	http.HandleFunc("/ir/stream", frameStreamHandler)
	http.HandleFunc("/ir/collector", collectorQueryHandler)
	http.HandleFunc("/ir/frame", frameQueryHandler)

	signalCollectorServer := http.Server{Addr: ":8080"}
	signalCollectorServer.RegisterOnShutdown(func() {
		log.Print("Shutting down server")
	})
	go func() {
		intr := make(chan os.Signal)
		signal.Notify(intr, os.Interrupt)
		<-intr
		signalCollectorServer.Shutdown(context.Background())
	}()
	log.Print("Server started")
	if err := signalCollectorServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
