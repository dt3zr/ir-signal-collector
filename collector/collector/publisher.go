package collector

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

type publishClient struct {
	serverURL string
}

func newPublishClient(serverHost string, serverPort int) (*publishClient, error) {
	serverURLString := fmt.Sprintf("http://%s:%d/ir/frame", serverHost, serverPort)
	_, err := url.Parse(serverURLString)
	if err == nil {
		return &publishClient{serverURLString}, nil
	}
	return nil, err
}

func (pc *publishClient) publishTaggedFrameJSON(taggedFrameJSON []byte) {
	response, err := http.Post(pc.serverURL, "application/json", bytes.NewReader(taggedFrameJSON))
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("Published frame. Response %v received\n", response.StatusCode)
}
