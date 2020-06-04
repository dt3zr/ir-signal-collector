package server

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func frameSimpleQueryHandler(response http.ResponseWriter, request *http.Request) {
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
}

func frameQueryHandler(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		response.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// parse the URL for collector id, protocol id, value
	var cid, pid, value string
	stringCount := scanURLParam("/signal/", request.URL.Path, &cid, &pid, &value)
	log.Printf("Found %d params from %s. Namely '%s', '%s', '%s'", stringCount, request.URL.Path, cid, pid, value)
	if stringCount < 1 {
		response.WriteHeader(http.StatusBadRequest)
		return
	}
	var output []byte
	switch stringCount {
	case 1:
		// This case handles the pattern /signal/[collectorID]
		if cid != "" {
			protocolIDList, err := db.getProtocolIDList(cid)
			if err != nil {
				log.Print(err)
				response.WriteHeader(http.StatusNotFound)
				return
			}
			pidStringList := make([]string, 0, len(protocolIDList))
			for p := range protocolIDList {
				pidStringList = append(pidStringList, protocolID(p).String())
			}
			output, err = json.Marshal(pidStringList)
			if err != nil {
				log.Print(err)
				response.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	case 2:
		// This case handles the pattern /signal/[collectorID]/[protocolID]
		if cid != "" && pid != "" {
			var p protocolID
			p.parse(pid)
			values, err := db.getValues(cid, p)
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
		// This case handles the pattern /signal/[collectorID]/[protocolID]/[value]
		if cid != "" && pid != "" && value != "" {
			var p protocolID
			p.parse(pid)
			frames, err := db.getFrameList(cid, p, value)
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
}

func frameStreamHandler(w http.ResponseWriter, r *http.Request) {
	// c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"*.golang.org"}})
	c, err := websocket.Accept(w, r, nil)
	log.Printf("Accepted websocket request from %s", r.RemoteAddr)
	defer log.Printf("Closing websocket connection for %s", r.RemoteAddr)
	defer c.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var v interface{} = db
	notifier := v.(frameNotifier)
	onNewFrame := notifier.notify(getSubscriberID(r.RemoteAddr))
	timer := time.NewTimer(15 * time.Minute)
	timeUp := false
	for !timeUp {
		select {
		case f := <-onNewFrame:
			if err = wsjson.Write(ctx, c, f); err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case <-timer.C:
			timeUp = true
		}
	}
}

func getSubscriberID(data string) string {
	h := sha1.Sum([]byte(data))
	b := make([]byte, 0, 20)
	for _, v := range h {
		b = append(b, v)
	}
	return hex.EncodeToString(b)
}

func scanURLParam(prefix, path string, out ...*string) int {
	if !strings.HasSuffix(prefix, "/") {
		prefix = strings.Join([]string{prefix, "/"}, "")
	}
	urlPath := strings.TrimPrefix(path, prefix)
	n := 0
	for _, sp := range out {
		pos := strings.Index(urlPath, "/")
		if pos < 0 {
			if len(urlPath) > 0 {
				*sp = urlPath[:]
				n++
			}
			break
		}
		*sp = urlPath[:pos]
		n++
		urlPath = urlPath[(pos + 1):]
	}
	return n
}
