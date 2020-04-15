package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var store = frameStore{frameSet: make(map[frameID][][]markSpacePair)}

// Start function sets up the url mapping and launches the HTTP
// server to listen on port 8080
func Start() {
	http.HandleFunc("/signal", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		payload, err := ioutil.ReadAll(request.Body)
		if err != nil {
			fmt.Println(err)
			return
		}

		r := new(signalPublishRequest)

		if err := json.Unmarshal(payload, r); err != nil {
			fmt.Println(err)
			return
		}

		store.add(r.ProtocolName, r.Value, r.Header, r.RawPulses)
		fmt.Println(store)
	})

	signalCollectorServer := http.Server{Addr: "localhost:8080"}
	if err := signalCollectorServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
