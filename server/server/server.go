package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var db = newDatabase()

// Start function sets up the url mapping and launches the HTTP
// server to listen on port 8080
func Start() {
	http.HandleFunc("/signal", func(response http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:
			taggedFrameJSON, err := ioutil.ReadAll(request.Body)
			if err != nil {
				log.Println(err)
				return
			}
			var theTaggedFrame taggedFrame
			if err := json.Unmarshal(taggedFrameJSON, &theTaggedFrame); err != nil {
				log.Println(err)
				return
			}
			log.Printf("Unmarshalled -> %+v\n", theTaggedFrame)
			db.insert(theTaggedFrame)
		case http.MethodGet:
			// construct signal list as response
			collectorIDList, err := db.getCollectorIDList()
			if err != nil {
				log.Println(err)
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
			collector2protocol := make(simpleCollectorProtocolMap)
			for _, cid := range collectorIDList {
				protocolIDList, err := db.getProtocolIDList(cid)
				if err != nil {
					log.Println(err)
					response.WriteHeader(http.StatusInternalServerError)
					return
				}
				protocol2Value := make(simpleProtocolValueMap)
				for _, pid := range protocolIDList {
					values, err := db.getValues(cid, pid)
					if err != nil {
						log.Println(err)
						response.WriteHeader(http.StatusInternalServerError)
						return
					}
					l := make(simpleValueLengthList, 0, len(values))
					for _, value := range values {
						frames, err := db.getFrameList(cid, pid, value)
						if err != nil {
							log.Println(err)
							response.WriteHeader(http.StatusInternalServerError)
							return
						}
						f := simpleValueLength{value, len(frames)}
						l = append(l, f)
					}
					protocol2Value[pid.String()] = l
				}
				collector2protocol[cid] = protocol2Value
			}
			// encode response
			responseBytes, err := json.Marshal(collector2protocol)
			if err != nil {
				log.Print(err)
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

	// URL patterns: [
	//   /signal/collectorID
	//   /signal/collectorID/protocolID
	//	 /signal/collectorID/protocolID/value
	// ]
	http.HandleFunc("/signal/", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			response.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// parse the URL for collector id, protocol id, value
		urlPath := strings.TrimPrefix(request.URL.Path, "/signal/")
		var collectorID, pid, value string
		stringPtr := []interface{}{&collectorID, &pid, &value}
		stringCount := 0
		for _, sp := range stringPtr {
			pos := strings.Index(urlPath, "/")
			if pos < 0 {
				if len(urlPath) > 0 {
					s := sp.(*string)
					*s = urlPath[:]
					stringCount++
				}
				break
			}
			s := sp.(*string)
			*s = urlPath[:pos]
			stringCount++
			urlPath = urlPath[(pos + 1):]
		}
		if stringCount < 1 {
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		var output []byte
		switch stringCount {
		case 1:
			if collectorID != "" {
				protocolIDList, err := db.getProtocolIDList(collectorID)
				if err != nil {
					log.Print(err)
					response.WriteHeader(http.StatusNotFound)
					return
				}
				output, err = json.Marshal(protocolIDList)
				if err != nil {
					log.Print(err)
					response.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		case 2:
			if collectorID != "" && pid != "" {
				var p protocolID
				p.parse(pid)
				values, err := db.getValues(collectorID, p)
				if err != nil {
					response.WriteHeader(http.StatusNotFound)
					return
				}
				output, err = json.Marshal(values)
				if err != nil {
					log.Print(err)
					response.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		case 3:
			if collectorID != "" && pid != "" && value != "" {
				var p protocolID
				p.parse(pid)
				frames, err := db.getFrameList(collectorID, p, value)
				if err != nil {
					log.Print(err)
					response.WriteHeader(http.StatusNotFound)
					return
				}
				output, err = json.Marshal(frames)
				if err != nil {
					log.Print(err)
					response.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
		response.Write(output)
	})

	signalCollectorServer := http.Server{Addr: "localhost:8080"}
	if err := signalCollectorServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
