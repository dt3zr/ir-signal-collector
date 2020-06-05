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

func frameStreamHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"localhost:*", "192.168.*.*:*"}})
	if err != nil {
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Accepted websocket request from %s", r.RemoteAddr)
	defer log.Printf("Closing websocket connection for %s", r.RemoteAddr)
	defer c.Close(websocket.StatusNormalClosure, "Handler exits")
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	db := <-dbLock
	var v interface{} = db
	notifier := v.(frameNotifier)
	onNewFrame, err := notifier.notify(getSubscriberID(r.RemoteAddr))
	dbUnlock <- db
	if err != nil {
		if debugMode {
			log.Print(err)
		}
		c.Close(websocket.StatusNormalClosure, "Already subscribed")
		return
	}
	timer := time.NewTimer(1 * time.Minute)
	for {
		select {
		case f := <-onNewFrame:
			if err = wsjson.Write(ctx, c, f); err != nil {
				log.Print(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case <-timer.C:
			db = <-dbLock
			v = db
			notifier = v.(frameNotifier)
			err = notifier.unNotify(getSubscriberID(r.RemoteAddr))
			if err != nil {
				if debugMode {
					log.Print(err)
				}
			}
			dbUnlock <- db
			return
		}
	}
}

func collectorQueryHandler(w http.ResponseWriter, r *http.Request) {
	db := <-dbLock
	defer func() { dbUnlock <- db }()
	collectorIDList, err := db.getCollectorIDList()
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	output, err := json.Marshal(collectorIDList)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(output)
}

func frameQueryHandler(w http.ResponseWriter, r *http.Request) {
	db := <-dbLock
	defer func() { dbUnlock <- db }()
	switch r.Method {
	case http.MethodPost:
		taggedFrameJSON, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading from request body.", err)
			return
		}
		var theTaggedFrame taggedFrame
		if err := json.Unmarshal(taggedFrameJSON, &theTaggedFrame); err != nil {
			log.Println("Error unmarshaling.", err)
			return
		}
		if debugMode {
			log.Printf("Unmarshalled -> %+v\n", theTaggedFrame)
		}
		db.insert(theTaggedFrame)
	case http.MethodGet:
		var collectorIDList []string
		c := r.URL.Query().Get("cid")
		// construct signal list as response
		if c != "" {
			collectorIDList = append(collectorIDList, c)
		} else {
			cList, err := db.getCollectorIDList()
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			collectorIDList = cList
		}
		collector2protocol := make(collector2ProtocolMap)
		for _, cid := range collectorIDList {
			protocolIDList, err := db.getProtocolIDList(cid)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			protocol2Value := make(protocol2ValueMap)
			for _, pid := range protocolIDList {
				values, err := db.getValues(cid, pid)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				value2FrameList := make(value2FrameListMap)
				for _, value := range values {
					frames, err := db.getFrameList(cid, pid, value)
					if err != nil {
						log.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					value2FrameList[value] = frames
				}
				protocol2Value[pid.String()] = value2FrameList
			}
			collector2protocol[cid] = protocol2Value
		}
		// encode response
		responseBytes, err := json.Marshal(collector2protocol)
		if err != nil {
			log.Print(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Write(responseBytes)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
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
