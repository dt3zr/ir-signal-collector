package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
)

var debugMode = true

var dbLock <-chan frameCRUD
var dbUnlock chan<- frameCRUD

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
	initDb := func(done <-chan interface{}) (<-chan frameCRUD, chan<- frameCRUD) {
		db := newDatabase()
		dbLock := make(chan frameCRUD)
		dbUnlock := make(chan frameCRUD)
		go func() {
			defer close(dbLock)
			defer close(dbUnlock)
			dbLock <- db
			for {
				select {
				case ref := <-dbUnlock:
					dbLock <- ref
				case <-done:
					return
				}
			}
		}()
		return dbLock, dbUnlock
	}
	done := make(chan interface{})
	dbLock, dbUnlock = initDb(done)
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
