package collector

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

func publishTaggedFrameJSON(taggedFrameJSON []byte) {
	response, err := http.Post("http://localhost:8080/signal", "application/json", bytes.NewReader(taggedFrameJSON))
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Printf("Published frame. Response %v received\n", response.StatusCode)
}
