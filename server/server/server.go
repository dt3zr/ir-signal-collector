package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var store = frameStore{frameSet: make(map[frameID][][]markSpacePair)}

// Start function sets up the url mapping and launches the HTTP
// server to listen on port 8080
func Start() {
	http.HandleFunc("/signal", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
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

		case http.MethodGet:
			// construct signal list as response
			r := signalListResponse{}
			for fid, frameList := range store.frameSet {
				f := frameMetadata{fid.protocolName, fid.value, len(frameList)}
				r.Metadata = append(r.Metadata, f)
			}

			// encode response
			responseBytes, err := json.Marshal(r)
			if err != nil {
				response.WriteHeader(http.StatusInternalServerError)
				return
			}

			response.Header().Add("Content-Type", "application/json")
			response.Header().Add("Access-Control-Allow-Origin", "*")
			response.Write(responseBytes)

		default:
			response.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/signal/", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// parse the URL for protocol name and value
		valuePath := strings.TrimPrefix(request.URL.Path, "/signal/")
		values := strings.Split(valuePath, "/")

		if len(values) != 2 || len(values[0]) == 0 || len(values[1]) == 0 {
			response.WriteHeader(http.StatusBadRequest)
			return
		}

		// obtain frames from frame store
		mspairList, ok := store.get(values[0], values[1])
		if !ok {
			response.WriteHeader(http.StatusNotFound)
			return
		}

		// construct frames from flattened frames
		frames := make([]frame, 0)
		for _, pair := range mspairList {
			head := markSpacePair{pair[0].Mark, pair[0].Space}
			f := frame{head, pair[2:]}
			frames = append(frames, f)
		}

		// encoding response
		r := signalQueryResponse{values[0], values[1], frames}
		responseString, err := json.Marshal(r)
		if err != nil {
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Write(responseString)
	})

	signalCollectorServer := http.Server{Addr: "localhost:8080"}
	if err := signalCollectorServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
